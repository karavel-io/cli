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

	"github.com/karavel-io/cli/pkg/logger"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

const HelmRepoName = "karavel"
const HelmDefaultRepo = "https://repository.platform.karavel.io"

var (
	ErrVersionEmpty = fmt.Errorf("version cannot be empty")
)

func NewRepo(version string, repoUrl string) (*repo.Entry, error) {
	if version == "" {
		return nil, ErrVersionEmpty
	}

	repoUrl = GetRepoUrl(version, repoUrl)

	name := HelmRepoName
	if version == "unstable" {
		name = UnstableRepoName()
	}

	return &repo.Entry{
		Name: name,
		URL:  repoUrl,
	}, nil
}

func WithRepository(ctx context.Context, entry *repo.Entry) (context.Context, error) {
	store := FromContext(ctx)
	if store.Has(entry.Name) {
		logger.FromContext(ctx).Debugf("repository name %q already exists", entry.Name)
		return ctx, nil
	}

	// Get settings
	settings := settingsFromContext(ctx)
	providers := getter.All(settings)

	// Initialize repo
	repo, err := repo.NewChartRepository(entry, providers)
	if err != nil {
		return ctx, err
	}

	// Use custom cache path to not affect the system installation
	repo.CachePath = settings.RepositoryCache

	// Try fetching index
	_, err = repo.DownloadIndexFile()
	if err != nil {
		return ctx, err
	}

	// Add to store
	store.Update(entry)

	// Write updated config to file
	err = store.WriteFile(settings.RepositoryConfig, 0644)
	if err != nil {
		return ctx, err
	}

	return withSettings(withStore(ctx, store), settings), nil
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
