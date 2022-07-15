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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortReleases(t *testing.T) {
	rels := []string{"2022.1", "2021.2", "2022.5-rc.1", "2022.5", "2022.5-rc.7", "2022.4-rc.2"}
	exp := []string{"2022.5", "2022.5-rc.7", "2022.5-rc.1", "2022.4-rc.2", "2022.1", "2021.2"}

	sortReleases(rels)
	assert.Equal(t, exp, rels)
}
