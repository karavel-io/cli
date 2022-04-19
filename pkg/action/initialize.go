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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/karavel-io/cli/internal/github"
	"github.com/karavel-io/cli/pkg/logger"
)

type InitParams struct {
	Workdir        string
	Filename       string
	KaravelVersion string
	Force          bool
	GitHubRepo     string
}

const cfgTpl = `version = "{{ . }}"

#  Now you can add some Karavel components to install in your cluster.
#  For a list of available components consult https://platform.karavel.io/components/
#
#  component "example" {
#    namespace = "example"
#
#    some = "param"
#    other = {
#      configuration = "values"
#    }
#  }
`

const (
	githubRepo      = "karavel-io/platform"
	githubApiURLTpl = "https://api.github.com/repos/%s/tags"
)

func Initialize(ctx context.Context, params InitParams) error {
	workdir := params.Workdir
	ver := params.KaravelVersion
	filename := params.Filename
	force := params.Force

	log := logger.FromContext(ctx)
	log.Infof("Initializing new Karavel %s project at %s", ver, workdir)
	log.Info()

	if err := os.MkdirAll(workdir, 0o755); err != nil {
		return err
	}

	filedst := filepath.Join(workdir, filename)
	info, err := os.Stat(filedst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if info != nil && !force {
		return fmt.Errorf("Karavel config file %s already exists", filename)
	}

	if info != nil && force {
		log.Warnf("Karavel config file %s already exists and will be overwritten", filename)
	}

	if ver == "" || ver == "latest" {
		repo := params.GitHubRepo
		if repo == "" {
			repo = githubRepo
		}

		apiUrl := fmt.Sprintf(githubApiURLTpl, repo)
		log.Infof("Fetching latest release version for GitHub repo %s", repo)
		ver, err = github.FetchLatestRelease(ctx, apiUrl)
		if err != nil {
			return err
		}
	}

	cfg, err := template.New("karavel.hcl").Parse(cfgTpl)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err = cfg.Execute(&buf, ver); err != nil {
		return err
	}

	log.Info()
	log.Infof("Writing config file to %s", filedst)
	return ioutil.WriteFile(filedst, buf.Bytes(), 0o655)
}
