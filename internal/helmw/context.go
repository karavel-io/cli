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
	"os"
	"path/filepath"

	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/repo"
)

type contextKey string

var (
	repoKey     contextKey = "helmw.repo"
	settingsKey contextKey = "helmw.settings"
)

func FromContext(ctx context.Context) *repo.File {
	val, ok := ctx.Value(repoKey).(*repo.File)
	if !ok || val == nil {
		return repo.NewFile()
	}
	return val
}

func withStore(ctx context.Context, store *repo.File) context.Context {
	return context.WithValue(ctx, repoKey, store)
}

func settingsFromContext(ctx context.Context) *cli.EnvSettings {
	val, ok := ctx.Value(settingsKey).(*cli.EnvSettings)
	if !ok || val == nil {
		settings := cli.New()
		settings.RepositoryCache = tmpdirfallback("helmcache")
		settings.RepositoryConfig = tmpfilefallback("helmconfig")
		settings.Debug = true
		return settings
	}
	return val
}

func withSettings(ctx context.Context, settings *cli.EnvSettings) context.Context {
	return context.WithValue(ctx, settingsKey, settings)
}

func tmpdirfallback(name string) string {
	// Try making temporary directory
	tmpdir, err := os.MkdirTemp("", name+"-")
	if err != nil {
		// Use dummy fallback
		return filepath.Join(os.TempDir(), name)
	}
	return tmpdir
}

func tmpfilefallback(name string) string {
	// Try making temporary directory
	file, err := os.CreateTemp("", name+"-")
	if err != nil {
		// Use dummy fallback
		return filepath.Join(os.TempDir(), name)
	}
	defer file.Close()
	return file.Name()
}
