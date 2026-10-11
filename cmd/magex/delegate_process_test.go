package main

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errNoProcessList = errors.New("no process list")

// TestParseProcessList checks that ps output is read into PIDs keyed by their
// parent's PID, skipping lines that aren't a PID pair.
func TestParseProcessList(t *testing.T) {
	out := "    1     0\n  100     1\n  101   100\n  102   100\n\n garbage line here\n  abc   100\n  103   101\n  104   104\n"

	got := parseProcessList(out)

	assert.Equal(t, map[int][]int{
		0:   {1},
		1:   {100},
		100: {101, 102},
		101: {103},
	}, got)
}

// TestDescendantPIDs checks that every process below a PID is found, parents
// before their children, and that a cycle in the list can't loop forever.
func TestDescendantPIDs(t *testing.T) {
	children := map[int][]int{
		1:   {100, 200},
		100: {101, 102},
		101: {103},
		200: {201},
		// A PID reused while ps listed the processes could make a cycle
		103: {100},
	}

	assert.Equal(t, []int{101, 102, 103}, descendantPIDs(children, 100))
	assert.Equal(t, []int{201}, descendantPIDs(children, 200))
	assert.Empty(t, descendantPIDs(children, 999))
	assert.Empty(t, descendantPIDs(nil, 100))
}

// TestDelegateOptionsProcessTree checks that a failed process list gives an
// empty one, since reaching the command's processes is a best effort.
func TestDelegateOptionsProcessTree(t *testing.T) {
	failing := delegateOptions{processes: func() (map[int][]int, error) { return nil, errNoProcessList }}
	assert.Nil(t, failing.processTree())

	fixed := delegateOptions{processes: func() (map[int][]int, error) { return map[int][]int{1: {2}}, nil }}
	assert.Equal(t, map[int][]int{1: {2}}, fixed.processTree())
}

// TestProcessChildren checks that the real process list includes this test
// process under its parent.
func TestProcessChildren(t *testing.T) {
	if runtime.GOOS == "windows" {
		_, err := processChildren()
		require.ErrorIs(t, err, errProcessListUnsupported)
		return
	}
	if _, err := exec.LookPath("ps"); err != nil {
		t.Skip("ps is not installed")
	}

	children, err := processChildren()
	require.NoError(t, err)
	assert.Contains(t, children[os.Getppid()], os.Getpid())
}
