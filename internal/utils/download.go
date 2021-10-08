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

package utils

import (
	"context"
	"github.com/karavel-io/cli/pkg/logger"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"os"
	"path"
	"time"
)

// Thank you Cosmin!
// https://gist.github.com/albulescu/e61979cc852e4ee8f49c

func DownloadWithProgress(ctx context.Context, log logger.Logger, url string, filename string) error {
	file := path.Base(url)

	log.Infof("Downloading file %s from %s\n", file, url)
	log.Info()

	start := time.Now()

	out, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	defer out.Close()

	headResp, err := http.Head(url)
	if err != nil {
		return err
	}

	defer headResp.Body.Close()
	if headResp.StatusCode >= 400 {
		return errors.Errorf("failed to fetch %s: %s", url, headResp.Status)
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	elapsed := time.Since(start)
	log.Info()
	log.Infof("Download completed in %s", elapsed)
	return nil
}
