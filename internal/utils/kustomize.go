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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
	"modernc.org/sortutil"
	"sigs.k8s.io/kustomize/api/types"
)

func RenderKustomizeFile(outdir string, resources []string, ignoreFn func(s string) bool) error {
	filename := filepath.Join(outdir, "kustomization.yml")
	exists := false
	info, err := os.Stat(filename)
	if err == nil || !os.IsNotExist(err) {
		exists = true
	}

	if info != nil && info.IsDir() {
		return fmt.Errorf("could not render %s: is a directory", filename)
	}

	var kfile types.Kustomization
	if exists {
		f, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}

		if err := kfile.Unmarshal(f); err != nil {
			return err
		}
		rr := make([]string, 0)
		for _, r := range kfile.Resources {
			if !ignoreFn(r) {
				rr = append(rr, r)
			}
		}

		resources = append(resources, rr...)
		resources = resources[:sortutil.Dedupe(sort.StringSlice(resources))]
	}

	kfile.FixKustomizationPostUnmarshalling()

	kfile.Resources = resources

	kfile.FixKustomizationPreMarshalling()

	// Cleanup paths and make sure they're using unix-style separators
	// This is mostly useful for Windows, other OSs shouldn't worry about this
	for index, resource := range kfile.Resources {
		resource = filepath.ToSlash(filepath.Clean(resource))
		kfile.Resources[index] = resource
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&kfile); err != nil {
		return err
	}

	if err := enc.Close(); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filename, buf.Bytes(), 0o655); err != nil {
		return err
	}

	return nil
}
