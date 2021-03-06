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

package config

import (
	"errors"
	"io"
	"strings"

	"github.com/karavel-io/cli/internal/helmw"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/json"
)

var ErrConfigParseFailed = errors.New("failed to parse Karavel config")

type Config struct {
	Version             string      `hcl:"version"`
	Components          []Component `hcl:"component,block"`
	HelmStableRepoUrl   string      `hcl:"stable_repo,optional"`
	HelmUnstableRepoUrl string      `hcl:"unstable_repo,optional"`
}

func ReadFrom(logw io.Writer, filename string) (Config, error) {
	var c Config

	p := hclparse.NewParser()

	w := hcl.NewDiagnosticTextWriter(logw, p.Files(), 79, true)
	f, err := p.ParseHCLFile(filename)
	if err != nil {
		_ = w.WriteDiagnostics(err)
		if err.HasErrors() {
			return c, ErrConfigParseFailed
		}
	}

	if err := gohcl.DecodeBody(f.Body, nil, &c); err != nil {
		_ = w.WriteDiagnostics(err)
		if err.HasErrors() {
			return c, ErrConfigParseFailed
		}
	}

	for i := range c.Components {
		cc := &c.Components[i]
		cc.Name = strings.ToLower(cc.Name)

		pp := make(map[string]cty.Value)
		for l, a := range cc.RawParams {
			v, err := a.Expr.Value(nil)
			if err != nil {
				_ = w.WriteDiagnostics(err)
				if err.HasErrors() {
					return c, ErrConfigParseFailed
				}
			}
			pp[l] = v
		}
		m := cty.ObjectVal(pp)
		j, jerr := json.Marshal(m, m.Type())
		if jerr != nil {
			return c, jerr
		}

		if strings.HasPrefix(strings.ToLower(cc.Version), "unstable:") {
			cc.Version = strings.SplitAfter(cc.Version, ":")[1]
			cc.Unstable = true
		}
		cc.JsonParams = string(j)
	}

	c.HelmStableRepoUrl = helmw.GetRepoUrl(c.Version, c.HelmStableRepoUrl)
	c.HelmUnstableRepoUrl = helmw.GetRepoUrl("unstable", c.HelmUnstableRepoUrl)

	return c, nil
}
