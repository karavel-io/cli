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
