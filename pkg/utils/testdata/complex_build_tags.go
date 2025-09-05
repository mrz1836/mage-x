//go:build integration && !windows
// +build integration,!windows

package testdata

import "testing"

func TestComplexBuildTags(t *testing.T) {
	t.Log("Complex build tags test")
}
