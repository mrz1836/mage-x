package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHasHelpFlag verifies the agent-footgun guard: any of -h, --help, -help
// among the post-command args should be detected so the CLI can show the
// command's help page instead of executing the command with the flag silently
// ignored as an unknown parameter.
func TestHasHelpFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "empty args", args: nil, want: false},
		{name: "no help flag", args: []string{"name=TestFoo", "pkg=./pkg/mage"}, want: false},
		{name: "value containing dashes is not a flag", args: []string{"name=Test-Foo"}, want: false},
		{name: "--help present", args: []string{"--help"}, want: true},
		{name: "-h present", args: []string{"-h"}, want: true},
		{name: "-help (single-dash) present", args: []string{"-help"}, want: true},
		{name: "--help after real params", args: []string{"name=TestFoo", "--help"}, want: true},
		{name: "-h before real params", args: []string{"-h", "pkg=./pkg/utils"}, want: true},
		{name: "similar but not exact: --helpful", args: []string{"--helpful"}, want: false},
		{name: "param that happens to mention help", args: []string{"name=TestHelp"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, hasHelpFlag(tt.args))
		})
	}
}
