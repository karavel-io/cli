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

package plan

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/karavel-io/cli/internal/argo"
	"github.com/karavel-io/cli/internal/helmw"
	"github.com/karavel-io/cli/internal/utils"
	"github.com/karavel-io/cli/internal/utils/predicate"
	"github.com/karavel-io/cli/pkg/logger"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart"
)

const (
	bootstrapAnnotation    = "karavel.io/bootstrap"
	dependenciesAnnotation = "karavel.io/dependencies"
	singletonAnnotation    = "karavel.io/singleton"
)

// reserved annotations that should not be interpreted as
// feature flags
var reservedAnnotations = map[string]bool{
	bootstrapAnnotation:    true,
	dependenciesAnnotation: true,
}

type Component struct {
	name             string
	component        string
	namespace        string
	version          string
	bootstrap        bool
	singleton        bool
	dependencies     []string
	integrationsDeps map[string][]string
	integrations     map[string]bool
	jsonParams       string
	unstable         bool
}

func NewComponentFromChartMetadata(meta *chart.Metadata, unstable bool) (Component, error) {
	var deps []string
	depsCsv := meta.Annotations[dependenciesAnnotation]
	if depsCsv != "" {
		depsCsv = strings.ReplaceAll(depsCsv, " ", "")
		cr := csv.NewReader(strings.NewReader(depsCsv))
		d, err := cr.Read()
		if err != nil {
			return Component{}, err
		}
		deps = d
	}

	integs := make(map[string][]string)
	for integ, reqsCsv := range meta.Annotations {
		if reservedAnnotations[integ] {
			continue
		}

		reqsCsv = strings.ReplaceAll(reqsCsv, " ", "")
		cr := csv.NewReader(strings.NewReader(reqsCsv))
		reqs, err := cr.Read()
		if err != nil {
			return Component{}, fmt.Errorf("failed to read integration dependencies: %w", err)
		}
		integs[integ] = reqs
	}

	bootstrap, err := strconv.ParseBool(meta.Annotations[bootstrapAnnotation])
	if err != nil {
		bootstrap = false
	}

	singleton, err := strconv.ParseBool(meta.Annotations[singletonAnnotation])
	if err != nil {
		singleton = false
	}

	return Component{
		name:             meta.Name,
		component:        meta.Name,
		version:          meta.Version,
		bootstrap:        bootstrap,
		singleton:        singleton,
		dependencies:     deps,
		integrationsDeps: integs,
		unstable:         unstable,
	}, nil
}

func (c *Component) Name() string {
	return c.name
}

func (c *Component) ComponentName() string {
	return c.component
}

func (c *Component) NameOverride() string {
	if c.component != c.name {
		return c.name
	}

	return ""
}

func (c *Component) Namespace() string {
	return c.namespace
}

func (c *Component) Version() string {
	return c.version
}

func (c *Component) IsBootstrap() bool {
	return c.bootstrap
}

func (c *Component) Params() string {
	return c.jsonParams
}

func (c *Component) GetParam(path string) gjson.Result {
	return gjson.Get(c.jsonParams, path)
}

func (c *Component) DebugLabel() string {
	var withAlias string
	if name := c.NameOverride(); name != "" {
		withAlias = fmt.Sprintf(" with alias '%s'", name)
	}
	return fmt.Sprintf("'%s' %s%s", c.ComponentName(), c.Version(), withAlias)
}

type routineRes struct {
	filename string
	err      error
}

func (c *Component) Render(ctx context.Context, log logger.Logger, outdir string) error {
	const errorFormat = "failed to render component '%s' v%s: %w"

	if err := os.RemoveAll(outdir); err != nil {
		return fmt.Errorf(errorFormat, c.name, c.version, err)
	}

	if err := os.MkdirAll(outdir, 0o755); err != nil {
		return fmt.Errorf(errorFormat, c.name, c.version, err)
	}

	if no := c.NameOverride(); no != "" {
		j, err := sjson.Set(c.jsonParams, "nameOverride", no)
		if err != nil {
			return fmt.Errorf(errorFormat, c.name, c.version, err)
		}

		c.jsonParams = j
	}

	docs, err := helmw.TemplateChart(ctx, c.component, helmw.ChartOptions{
		Namespace: c.namespace,
		Version:   c.version,
		Values:    c.jsonParams,
		Unstable:  c.unstable,
	})
	if err != nil {
		return fmt.Errorf(errorFormat, c.name, c.version, err)
	}

	log.Debugf("component %s: writing %d resources", c.DebugLabel(), len(docs))

	var wg sync.WaitGroup
	resch := make(chan routineRes)

	for i, d := range docs {
		if len(d) == 0 {
			continue
		}

		wg.Add(1)
		go func(i int, doc helmw.YamlDoc) {
			defer wg.Done()

			var buf bytes.Buffer
			enc := yaml.NewEncoder(&buf)
			enc.SetIndent(2)
			if err := enc.Encode(&doc); err != nil {
				resch <- routineRes{err: err}
				return
			}

			if err := enc.Close(); err != nil {
				resch <- routineRes{err: err}
				return
			}

			k := strings.ToLower(doc["kind"].(string))
			meta := doc["metadata"].(helmw.YamlDoc)
			ns := ""
			if meta["namespace"] != nil && meta["namespace"] != c.namespace {
				ns = "-" + meta["namespace"].(string)
			}

			basename := fmt.Sprintf("%s-%s%s.yml", k, meta["name"], ns)
			filename := filepath.Join(outdir, basename)
			log.Debugf("component %s writing file %s", c.DebugLabel(), filepath.Join(filepath.Base(outdir), basename))
			if err := ioutil.WriteFile(filename, buf.Bytes(), 0o655); err != nil {
				resch <- routineRes{err: err}
			}

			resch <- routineRes{filename: filepath.Base(basename)}
		}(i, d)
	}

	go func() {
		wg.Wait()
		close(resch)
	}()

	var files []string
	for res := range resch {
		if res.err != nil {
			return fmt.Errorf(errorFormat, c.name, c.version, res.err)
		}
		files = append(files, res.filename)
	}

	sort.Strings(files)
	return utils.RenderKustomizeFile(outdir, files, predicate.StringFalse)
}

func (c *Component) RenderApplication(argoNs string, repoUrl string, path string, outfile string) error {
	app := argo.NewApplication(c.name, c.namespace, argoNs, repoUrl, path)
	return app.Render(outfile)
}

func (c *Component) patchIntegrations(log logger.Logger) error {
	jp := c.jsonParams
	log.Debugf("Processing integrations for component '%s'", c.Name())
	for param, status := range c.integrations {
		log.Debugf("Processing integration %s: %t for component '%s'", param, status, c.Name())
		curr := gjson.Get(jp, param)
		if curr.Exists() {
			log.Debugf("Override for integration %s is present for component '%s'. Skipping integration", param, c.Name())
		} else {
			log.Debugf("Override for integration %s not present for component '%s'. Setting to %t", param, c.Name(), status)
			j, err := sjson.Set(jp, param, status)
			if err != nil {
				return err
			}
			jp = j
		}
	}
	c.jsonParams = jp
	return nil
}
