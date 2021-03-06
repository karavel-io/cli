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

package argo

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v3"
)

// Application is a lightweight struct matching argoproj.io/v1alpha1/Application
type Application struct {
	TypeMeta   `yaml:",inline"`
	ObjectMeta `yaml:"metadata"`
	Spec       ApplicationSpec `yaml:"spec"`
}

type ApplicationSpec struct {
	Source Source `yaml:"source"`
	// Destination overrides the kubernetes server and namespace defined in the environment ksonnet app.yaml
	Destination Destination `yaml:"destination"`
	Project     string      `yaml:"project"`
	SyncPolicy  SyncPolicy  `yaml:"syncPolicy,omitempty"`
}

type Source struct {
	RepoUrl string `yaml:"repoURL"`
	Path    string `yaml:"path"`
}

type Destination struct {
	Server    string `yaml:"server"`
	Namespace string `yaml:"namespace"`
}

type SyncPolicy struct {
	Automated   Automated `yaml:"automated"`
	SyncOptions []string  `yaml:"syncOptions"`
	Retry       Retry     `yaml:"retry"`
}

type Automated struct {
	Prune      bool `yaml:"prune"`
	SelfHeal   bool `yaml:"selfHeal"`
	AllowEmpty bool `yaml:"allowEmpty"`
}

type Retry struct {
	Limit   int     `yaml:"limit"`
	Backoff Backoff `yaml:"backoff"`
}

type Backoff struct {
	Duration    time.Duration `yaml:"duration"`
	Factor      int           `yaml:"factor"`
	MaxDuration time.Duration `yaml:"maxDuration"`
}

func NewApplication(name string, namespace string, argoNs string, repoUrl string, path string) Application {
	return Application{
		TypeMeta: TypeMeta{
			APIVersion: "argoproj.io/v1alpha1",
			Kind:       "Application",
		},
		ObjectMeta: ObjectMeta{
			Name:      name,
			Namespace: argoNs,
			Annotations: map[string]string{
				"argocd.argoproj.io/manifest-generate-paths": ".",
			},
		},
		Spec: ApplicationSpec{
			Source: Source{
				RepoUrl: repoUrl,
				Path:    path,
			},
			Destination: Destination{
				Server:    "https://kubernetes.default.svc",
				Namespace: namespace,
			},
			Project: "infrastructure",
			SyncPolicy: SyncPolicy{
				Automated: Automated{
					Prune:      true,
					SelfHeal:   true,
					AllowEmpty: false,
				},
				SyncOptions: []string{
					"Validate=false",
					"CreateNamespace=true",
					"ApplyOutOfSyncOnly=true",
					"SkipDryRunOnMissingResource=true",
				},
				Retry: Retry{
					Limit: 5,
					Backoff: Backoff{
						Duration:    5 * time.Second,
						Factor:      2,
						MaxDuration: 3 * time.Minute,
					},
				},
			},
		},
	}
}

func (app *Application) Render(outfile string) error {
	deferr := fmt.Sprintf("failed to render application manifest '%s'", app.Name)
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&app); err != nil {
		return fmt.Errorf("%s: %w", deferr, err)
	}

	if err := enc.Close(); err != nil {
		return fmt.Errorf("%s: %w", deferr, err)
	}

	if err := ioutil.WriteFile(outfile, buf.Bytes(), 0o655); err != nil {
		return fmt.Errorf("%s: %w", deferr, err)
	}
	return nil
}
