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

import "github.com/hashicorp/hcl/v2"

type Component struct {
	Name          string                    `hcl:"name,label"`
	ComponentName string                    `hcl:"component,optional"`
	Namespace     string                    `hcl:"namespace,optional"`
	Version       string                    `hcl:"version,optional"`
	RawParams     map[string]*hcl.Attribute `hcl:",remain"`
	JsonParams    string
	Unstable      bool
}
