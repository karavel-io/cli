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

package helmw

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/karavel-io/cli/pkg/logger"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
)

func cliSettings() *cli.EnvSettings {
	settings := cli.New()
	settings.RepositoryCache = filepath.Join(os.TempDir(), "helmcache")
	settings.RepositoryConfig = filepath.Join(os.TempDir(), "helmconfig")
	return settings
}

func GetChartManifest(chartName string, version string, unstable bool) (*chart.Metadata, error) {
	repo := HelmRepoName
	if unstable {
		repo = UnstableRepoName()
	}

	chartName = fmt.Sprintf("%s/%s", repo, chartName)
	hshow := action.NewShow(action.ShowAll)
	hshow.Devel = unstable
	hshow.Version = version

	path, err := hshow.LocateChart(chartName, cliSettings())
	if err != nil {
		return nil, err
	}

	ch, err := loader.Load(path)
	if err != nil {
		return nil, err
	}

	return ch.Metadata, nil
}

type YamlDoc map[string]interface{}

type ChartOptions struct {
	Namespace string
	Version   string
	Values    string
	Unstable  bool
}

func TemplateChart(ctx context.Context, name string, options ChartOptions) ([]YamlDoc, error) {
	repo := HelmRepoName
	if options.Unstable {
		repo = UnstableRepoName()
	}

	chartName := fmt.Sprintf("%s/%s", repo, name)

	settings := cliSettings()
	//providers := getter.All(settings)
	logger := logger.FromContext(ctx)

	config := new(action.Configuration)
	err := config.Init(
		settings.RESTClientGetter(),
		settings.Namespace(),
		"", // defaults to secret
		logger.Debugf,
	)
	if err != nil {
		return nil, fmt.Errorf("could not initialize config object: %w", err)
	}

	// Create install action for generating charts
	install := action.NewInstall(config)

	// Don't actually install
	install.DryRun = true
	install.IncludeCRDs = true
	install.ClientOnly = true
	install.SkipCRDs = false
	install.Replace = true
	install.ReleaseName = "dummy"

	// Copy values from options
	install.Version = options.Version
	install.Namespace = options.Namespace

	// Get chart
	chartPath, err := install.LocateChart(chartName, settings)
	if err != nil {
		return nil, fmt.Errorf("could not locate chart: %w", err)
	}

	// Load chart
	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("could not load chart: %w", err)
	}

	// Get values
	var values map[string]any
	err = yaml.Unmarshal([]byte(options.Values), &values)
	if err != nil {
		return nil, fmt.Errorf("could not decode values: %w", err)
	}

	// Run install and generate manifests
	release, err := install.Run(chart, values)
	if err != nil {
		return nil, fmt.Errorf("helm install failed: %w", err)
	}
	if release == nil {
		logger.Warnf("chart \"%s\" has generated no manifests", name)
		return []YamlDoc{}, nil
	}

	// Decode generated manifests
	dec := yaml.NewDecoder(strings.NewReader(release.Manifest))
	var docs []YamlDoc
	for {
		var doc YamlDoc
		if err := dec.Decode(&doc); err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		docs = append(docs, doc)
	}
	return docs, nil
}
