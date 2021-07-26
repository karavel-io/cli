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
	"bytes"
	"fmt"
	helmclient "github.com/mittwald/go-helm-client"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"io"
)

func GetChartManifest(chartName string, version string, unstable bool) (*chart.Metadata, error) {
	repo := HelmRepoName
	if unstable {
		repo = UnstableRepoName()
	}

	chartName = fmt.Sprintf("%s/%s", repo, chartName)
	hc, err := helmclient.New(&helmclient.Options{
		Debug: true,
	})
	if err != nil {
		return nil, err
	}
	h := hc.(*helmclient.HelmClient)
	hshow := action.NewShow(action.ShowAll)
	hshow.Devel = unstable
	hshow.Version = version

	path, err := hshow.LocateChart(chartName, h.Settings)
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

func TemplateChart(name string, namespace string, version string, values string, unstable bool) ([]YamlDoc, error) {
	repo := HelmRepoName
	if unstable {
		repo = UnstableRepoName()
	}
	hc, err := helmclient.New(&helmclient.Options{
		Debug: true,
	})
	if err != nil {
		return nil, err
	}
	h := hc.(*helmclient.HelmClient)

	ch := &helmclient.ChartSpec{
		ChartName:  fmt.Sprintf("%s/%s", repo, name),
		Namespace:  namespace,
		Version:    version,
		SkipCRDs:   false,
		ValuesYaml: values,
	}

	manifests, err := h.TemplateChart(ch)
	if err != nil {
		return nil, err
	}

	dec := yaml.NewDecoder(bytes.NewReader(manifests))

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
