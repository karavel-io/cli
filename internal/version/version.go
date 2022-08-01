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

package version

import (
	"runtime"
	"runtime/debug"
	"time"
)

var (
	version       = ""
	gitCommit     = ""
	gitTreeState  = ""
	gitCommitDate = ""
	buildDate     = time.Now().UTC().Format(time.RFC3339)
)

type Version struct {
	Version       string
	GitCommit     string
	GitTreeState  string
	GitCommitDate string
	BuildDate     string
	Arch          string
	Os            string
	GoVersion     string
}

func Get() Version {
	info, ok := debug.ReadBuildInfo()
	if ok {
		if version == "" {
			version = info.Main.Version
		}
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				gitCommit = setting.Value
			case "vcs.time":
				gitCommitDate = setting.Value
			case "vcs.modified":
				if setting.Value == "true" {
					gitTreeState = "dirty"
				} else {
					gitTreeState = "clean"
				}
			}
		}
	}

	return Version{
		Version:       version,
		GitCommit:     gitCommit,
		GitTreeState:  gitTreeState,
		GitCommitDate: gitCommitDate,
		BuildDate:     buildDate,
		Arch:          runtime.GOARCH,
		Os:            runtime.GOOS,
		GoVersion:     runtime.Version(),
	}
}

func Short() string {
	ver := version
	if gitTreeState != "clean" {
		ver += "-" + gitTreeState
	}
	return ver
}
