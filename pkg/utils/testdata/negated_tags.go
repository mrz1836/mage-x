//go:build !nocgo && (unit || integration)
// +build !nocgo
// +build unit integration

package testdata

import "testing"

func TestNegatedTags(t *testing.T) {
	t.Log("Negated tags test")
}
