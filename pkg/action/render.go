// Copyright 2021 The Karavel Project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package action

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	argocd "github.com/karavel-io/cli/internal/argo"
	"github.com/karavel-io/cli/internal/gitutils"
	"github.com/karavel-io/cli/internal/helmw"
	"github.com/karavel-io/cli/internal/plan"
	"github.com/karavel-io/cli/internal/utils"
	"github.com/karavel-io/cli/internal/utils/predicate"
	"github.com/karavel-io/cli/pkg/config"
	"github.com/karavel-io/cli/pkg/logger"
)

type RenderParams struct {
	ConfigPath string
	SkipGit    bool
}

func addRepo(ctx context.Context, version string, url string) (context.Context, error) {
	// Create repo entry
	entry, err := helmw.NewRepo(version, url)
	if err != nil {
		return ctx, err
	}

	// Add repository
	ctx, err = helmw.WithRepository(ctx, entry)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func Render(ctx context.Context, params RenderParams) error {
	cpath := params.ConfigPath
	skipGit := params.SkipGit
	workdir := filepath.Dir(cpath)
	vendorDir := filepath.Join(workdir, "vendor")
	appsDir := filepath.Join(workdir, "applications")
	projsDir := filepath.Join(workdir, "projects")
	argoEnabled := true

	log := logger.FromContext(ctx)
	log.Infof("Rendering new Karavel project with config file %s", cpath)

	log.Debug("Reading config file")
	cfg, err := config.ReadFrom(log.Writer(), cpath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	log.Debugf("Karavel Container Platform version %s", cfg.Version)
	log.Debugf("Updating Karavel components stable repository %s", cfg.HelmStableRepoUrl)
	ctx, err = addRepo(ctx, cfg.Version, cfg.HelmStableRepoUrl)
	if err != nil {
		return fmt.Errorf("failed to setup Karavel stable components repository: %w", err)
	}

	log.Debugf("Updating Karavel components unstable repository %s", cfg.HelmUntableRepoUrl)
	ctx, err = addRepo(ctx, "unstable", cfg.HelmUntableRepoUrl)
	if err != nil {
		return fmt.Errorf("failed to setup Karavel stable components repository: %w", err)
	}

	defer helmw.Clean(ctx)

	log.Debug("Creating render plan from config")
	p, err := plan.NewFromConfig(ctx, &cfg)
	if err != nil {
		return fmt.Errorf("failed to instantiate render plan from config: %w", err)
	}

	log.Debug("Validating render plan")
	if err := p.Validate(); err != nil {
		return err
	}

	argo := p.GetComponent("argocd")
	if argo == nil {
		argoEnabled = false
		log.Warnf("ArgoCD component is missing. GitOps integrations will be disabled")
	}

	assertDirs := []string{vendorDir}
	if argoEnabled {
		assertDirs = append(assertDirs, appsDir, projsDir)
	}

	for _, dir := range assertDirs {
		log.Debugf("Asserting directory %s", dir)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	var wg sync.WaitGroup
	ch := make(chan utils.Pair)
	appNames := make(chan string)
	done := make(chan bool)

	var apps []string
	var renderDirs []string
	if argoEnabled {
		renderDirs = []string{"applications", "projects"}
	}
	dirInfos, err := ioutil.ReadDir(vendorDir)
	if err != nil {
		return err
	}

	dirs := make(map[string]struct{}, len(dirInfos))
	for _, i := range dirInfos {
		dirs[i.Name()] = struct{}{}
	}

	repoPath, repoUrl := "", ""
	if !skipGit && argoEnabled {
		res := argo.GetParam("git.repo")
		log.Debug("Finding remote git repository URL to configure ArgoCD applications")
		dir, url, err := gitutils.GetOriginRemote(log, workdir, res.String())
		if err != nil {
			return err
		}

		file, err := filepath.Rel(dir, workdir)
		if err != nil {
			return err
		}

		repoPath, repoUrl = file, url
	}

	// empty line for nice logs
	log.Info()

	for _, c := range p.Components() {
		if c.IsBootstrap() {
			renderDirs = append(renderDirs, filepath.Join("vendor", c.Name()))
		}
		delete(dirs, c.Name())

		wg.Add(1)
		go func(comp *plan.Component) {
			defer wg.Done()

			msg := fmt.Sprintf("failed to render component '%s'", comp.Name())
			outdir := filepath.Join(vendorDir, comp.Name())
			log.Infof("Rendering component %s at %s", comp.DebugLabel(), strings.ReplaceAll(outdir, filepath.Dir(workdir)+"/", ""))
			log.Debugf("Component %s params: %s", comp.DebugLabel(), comp.Params())

			if err := comp.Render(ctx, log, outdir); err != nil {
				ch <- utils.NewPair(msg, err)
				return
			}

			if argoEnabled {
				log.Debugf("Rendering application manifest for component %s", comp.DebugLabel())
				appFile := comp.Name() + ".yml"
				appNames <- appFile
				appFullPath := filepath.Join(appsDir, appFile)
				// if the application file already exists, we skip it. It has already been created
				// and we don't want to overwrite any changes the user may have made
				_, err = os.Stat(appFullPath)
				if !os.IsNotExist(err) {
					ch <- utils.NewPair(msg, err)
					return
				}

				argoNs := argo.Namespace()
				vendorPath := path.Join(repoPath, "vendor", comp.Name())
				if err := comp.RenderApplication(argoNs, repoUrl, vendorPath, appFullPath); err != nil {
					ch <- utils.NewPair(msg, err)
				}
			}
		}(c)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for dir := range dirs {
			log.Debugf("deleting extraneous directory '%s' in vendor", dir)
			if err := os.RemoveAll(filepath.Join(vendorDir, dir)); err != nil {
				ch <- utils.NewPair(fmt.Sprintf("failed to delete extraneous directory '%s' in vendor", dir), err)
			}
		}
	}()

	go func() {
		wg.Wait()
		done <- true
	}()

	open := true
	for open {
		select {
		case name := <-appNames:
			apps = append(apps, name)
		case pair := <-ch:
			err := pair.ErrorB()
			if err != nil {
				return fmt.Errorf("%s: %w", pair.StringA(), err)
			}
		case <-done:
			open = false
		}
	}

	if argoEnabled {
		argoNs := argo.Namespace()
		apps = append(apps, "projects.yml", "bootstrap.yml")
		sort.Strings(apps)
		if err := utils.RenderKustomizeFile(appsDir, apps, predicate.IsStringInSlice(apps)); err != nil {
			return fmt.Errorf("failed to render applications kustomization.yml: %w", err)
		}

		infraProj := "infrastructure.yml"
		if err := ioutil.WriteFile(filepath.Join(projsDir, infraProj), []byte(fmt.Sprintf(argoProject, argoNs)), 0o655); err != nil {
			return fmt.Errorf("failed to render infrastructure project file: %w", err)
		}

		projs := []string{infraProj}
		if err := utils.RenderKustomizeFile(projsDir, projs, predicate.IsStringInSlice(projs)); err != nil {
			return fmt.Errorf("failed to render projects kustomization.yml: %w", err)
		}

		projsAppPath := filepath.Join(appsDir, "projects.yml")
		_, err = os.Stat(projsAppPath)
		if os.IsNotExist(err) {
			projsApp := argocd.NewApplication("projects", "argocd", "argocd", repoUrl, path.Join(repoPath, "projects"))
			if err := projsApp.Render(projsAppPath); err != nil {
				return fmt.Errorf("failed to render projects application: %w", err)
			}
		}

		bootstrapAppPath := filepath.Join(appsDir, "bootstrap.yml")
		_, err = os.Stat(bootstrapAppPath)
		if os.IsNotExist(err) {
			bootstrap := argocd.NewApplication("bootstrap", "argocd", "argocd", repoUrl, path.Join(repoPath, "applications"))
			if err := bootstrap.Render(bootstrapAppPath); err != nil {
				return fmt.Errorf("failed to render bootstrap application: %w", err)
			}
		}

	}

	if err := utils.RenderKustomizeFile(workdir, renderDirs, predicate.StringOr(predicate.IsStringInSlice(renderDirs), predicate.StringHasPrefix("vendor"))); err != nil {
		return fmt.Errorf("failed to render kustomization.yml: %w", err)
	}

	return nil
}

const argoProject = `
apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: infrastructure
  namespace: %s
spec:
  description: Platform infrastructure components
  sourceRepos:
    - '*'
  destinations:
    - namespace: '*'
      server: 'https://kubernetes.default.svc'
  clusterResourceWhitelist:
    - group: '*'
      kind: '*'

`
