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
	"fmt"
	helmclient "github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/repo"
)

const HelmRepoName = "karavel"
const HelmDefaultRepo = "https://repository.platform.karavel.io"

func SetupHelm(version string, repoUrl string) error {
	if version == "" {
		return fmt.Errorf("version cannot be empty")
	}

	h, err := helmclient.New(&helmclient.Options{})
	if err != nil {
		return err
	}

	repoUrl = GetRepoUrl(version, repoUrl)

	name := HelmRepoName
	if version == "unstable" {
		name = UnstableRepoName()
	}

	if err := h.AddOrUpdateChartRepo(repo.Entry{
		Name: name,
		URL:  repoUrl,
	}); err != nil {
		return err
	}

	return h.UpdateChartRepos()
}

func GetRepoUrl(version string, repoUrl string) string {
	if repoUrl == "" {
		repoUrl = fmt.Sprintf("%s/%s", HelmDefaultRepo, version)
	}

	return repoUrl
}

func UnstableRepoName() string {
	return fmt.Sprintf("%s-unstable", HelmRepoName)
}
