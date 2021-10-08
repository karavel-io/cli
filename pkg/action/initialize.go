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
	"github.com/karavel-io/cli/internal/utils"
	"github.com/karavel-io/cli/pkg/logger"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"
)

const (
	latestReleaseURL = "https://github.com/karavel-io/platform/releases/latest/download"
	releaseUrl       = "https://github.com/karavel-io/platform/releases/v%s/download"
)

type InitParams struct {
	Workdir         string
	Filename        string
	KaravelVersion  string
	Force           bool
	FileUrlOverride string
}

func Initialize(log logger.Logger, params InitParams) error {
	workdir := params.Workdir
	ver := params.KaravelVersion
	filename := params.Filename
	force := params.Force

	log.Infof("Initializing new Karavel v%s project at %s", ver, workdir)
	log.Info()

	var baseUrlStr string
	if ver == "latest" {
		baseUrlStr = latestReleaseURL
	} else {
		baseUrlStr = fmt.Sprintf(releaseUrl, ver)
	}

	baseUrl, err := url.Parse(baseUrlStr)
	if err != nil {
		return errors.Wrap(err, "failed to parse download URL")
	}

	log.Warnf("URL: %s", baseUrlStr)

	cfgUrl := params.FileUrlOverride
	if cfgUrl == "" {
		baseUrl.Path = path.Join(baseUrl.Path, filename)
		cfgUrl = baseUrl.String()
	}

	log.Infof("Fetching starting config from %s", cfgUrl)
	log.Info()

	if err := os.MkdirAll(workdir, 0755); err != nil {
		return err
	}

	filedst := filepath.Join(workdir, filename)
	info, err := os.Stat(filedst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if info != nil && !force {
		return errors.Errorf("Karavel config file %s already exists", filename)
	}

	if info != nil && force {
		log.Warnf("Karavel config file %s already exists and will be overwritten", filename)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cfg, err := download(ctx, log, cfgUrl)
	if err != nil {
		return err
	}

	log.Info()
	log.Infof("Writing config file to %s", filedst)
	return ioutil.WriteFile(filedst, cfg, 0655)
}

func download(ctx context.Context, log logger.Logger, url string) ([]byte, error) {
	f, err := ioutil.TempFile("", path.Base(url))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	defer os.Remove(f.Name())

	if err := utils.DownloadWithProgress(ctx, log, url, f.Name()); err != nil {
		return nil, err
	}
	return ioutil.ReadFile(f.Name())
}
