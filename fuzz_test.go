package main_test

import "testing"

func FuzzTest(f *testing.F) { f.Fuzz(func(t *testing.T, data []byte) { _ = data }) }
