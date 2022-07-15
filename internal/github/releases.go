// Copyright 2022 The Karavel Project
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

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

const (
	ErrSetup = "failed to construct HTTP request from GitHub API"
	ErrHttp  = "failed to fetch releases from GitHub API"
	ErrJson  = "failed to decode JSON response from GitHub API"
)

type ghError struct {
	Message string `json:"message"`
}

type tag struct {
	Name string `json:"name"`
}

func FetchLatestRelease(ctx context.Context, apiUrl string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", apiUrl, nil)
	if err != nil {
		return "", fmt.Errorf("%s: %w", ErrSetup, err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%s: %w", ErrHttp, err)
	}

	defer res.Body.Close()

	if res.StatusCode >= 400 {
		var gherr ghError
		if err := json.NewDecoder(res.Body).Decode(&gherr); err != nil {
			return "", err
		}
		return "", fmt.Errorf("%s: %w", ErrHttp, fmt.Errorf("%s", gherr.Message))
	}

	tags := make([]tag, 0)
	if err := json.NewDecoder(res.Body).Decode(&tags); err != nil {
		return "", fmt.Errorf("%s: %w", ErrJson, err)
	}

	versions := make([]string, len(tags))
	for _, tag := range tags {
		versions = append(versions, tag.Name)
	}

	sortReleases(versions)
	return versions[0], nil
}

func sortReleases(releases []string) {
	for i, rel := range releases {
		if !strings.Contains(rel, "-rc") {
			releases[i] = rel + "-zzz" // This is so that 2022.5 is sorted before 2022.5-rc.1
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(releases)))

	for i, rel := range releases {
		releases[i] = strings.TrimSuffix(rel, "-zzz")
	}
}
