//go:build legacy && !old
// +build legacy,!old

package testdata

import "testing"

func TestLegacyOnly(t *testing.T) {
	t.Log("Legacy build tags only")
}
