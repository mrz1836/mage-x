//go:build integration && !windows

package integration

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// serveMagefile has a Serve target that runs until interrupted, then takes 2
// seconds to shut down.
const serveMagefile = `//go:build mage

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Serve runs until interrupted, then shuts down gracefully (takes 2s).
func Serve() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Println("serve: running, press Ctrl-C to stop")
	<-ctx.Done()
	fmt.Println("serve: interrupt received, shutting down...")
	time.Sleep(2 * time.Second) // e.g. stop child processes, flush state
	fmt.Println("serve: shutdown complete")
	return nil
}
`

// steppedServeMagefile has a Serve target that runs until interrupted, then
// shuts down in four 1-second steps, each writing a line to stdout and one to
// stderr: 4 seconds, which mage allows a target to clean up in 5.
const steppedServeMagefile = `//go:build mage

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Serve runs until interrupted, then shuts down in four 1-second steps.
func Serve() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Println("serve: running, press Ctrl-C to stop")
	<-ctx.Done()
	fmt.Println("serve: interrupt received, shutting down...")
	for step := 1; step <= 4; step++ {
		time.Sleep(time.Second)
		fmt.Printf("serve: cleanup step %d/4 done (stdout)\n", step)
		fmt.Fprintf(os.Stderr, "serve: cleanup step %d/4 done (stderr)\n", step)
	}
	fmt.Println("serve: shutdown complete")
	return nil
}
`

// timedLine is a line of output and when it arrived.
type timedLine struct {
	at   time.Time
	text string
}

// lineRecorder collects a process's output, noting when each line arrives.
type lineRecorder struct {
	mu      sync.Mutex
	all     bytes.Buffer
	partial []byte
	lines   []timedLine
}

// Write records p, timestamping each line it completes.
func (r *lineRecorder) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	r.all.Write(p)
	r.partial = append(r.partial, p...)
	for {
		i := bytes.IndexByte(r.partial, '\n')
		if i < 0 {
			break
		}
		r.lines = append(r.lines, timedLine{at: now, text: string(r.partial[:i])})
		r.partial = r.partial[i+1:]
	}
	return len(p), nil
}

// String returns everything recorded so far.
func (r *lineRecorder) String() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.all.String()
}

// Since renders the recorded lines with their time relative to start.
func (r *lineRecorder) Since(start time.Time) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	var out strings.Builder
	for _, line := range r.lines {
		fmt.Fprintf(&out, "%+6.2fs  %s\n", line.at.Sub(start).Seconds(), line.text)
	}
	if len(r.partial) > 0 {
		fmt.Fprintf(&out, "         %s\n", r.partial)
	}
	return out.String()
}

// TestMagex_CustomTargetHandlesCtrlC proves that a Ctrl-C, sent to the whole
// process group as a terminal does, lets a custom magefile target finish its
// own shutdown, as it does under plain mage:
//   - magex waits for mage instead of killing it, so all of the target's
//     cleanup output appears, none of it lost to a closed stderr pipe
//   - magex exits 130 without reporting the target as failed
//   - nothing is left running in the process group once magex exits
func TestMagex_CustomTargetHandlesCtrlC(t *testing.T) {
	if !hasMageBinary() {
		t.Skip("Skipping: mage binary not installed (required for custom magefile execution)")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("Skipping: go not installed (mage needs it to compile the magefile)")
	}
	bin := buildMagexForSignalTest(t)

	tests := []struct {
		name     string
		magefile string
		want     []string
	}{
		{
			name:     "two-second shutdown",
			magefile: serveMagefile,
			want: []string{
				"serve: interrupt received, shutting down...",
				"serve: shutdown complete",
			},
		},
		{
			name:     "four cleanup steps on stdout and stderr",
			magefile: steppedServeMagefile,
			want: []string{
				"serve: interrupt received, shutting down...",
				"serve: cleanup step 1/4 done (stdout)",
				"serve: cleanup step 1/4 done (stderr)",
				"serve: cleanup step 2/4 done (stdout)",
				"serve: cleanup step 2/4 done (stderr)",
				"serve: cleanup step 3/4 done (stdout)",
				"serve: cleanup step 3/4 done (stderr)",
				"serve: cleanup step 4/4 done (stdout)",
				"serve: cleanup step 4/4 done (stderr)",
				"serve: shutdown complete",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, out := startMagexServe(t, bin, tt.magefile)
			pgid := cmd.Process.Pid

			signaled := time.Now()
			require.NoError(t, syscall.Kill(-pgid, syscall.SIGINT))

			exitCode, timedOut := waitOrTimeout(t, cmd, 30*time.Second)
			exited := time.Since(signaled)
			t.Logf("magex exited with %d at %+.2fs; output, timed from the SIGINT:\n%s", exitCode, exited.Seconds(), out.Since(signaled))

			require.False(t, timedOut, "magex did not exit within 30s of SIGINT")
			assert.Equal(t, 130, exitCode)
			output := out.String()
			for _, line := range tt.want {
				assert.Contains(t, output, line)
			}
			assert.NotContains(t, output, "custom command failed")
			assertGroupGone(t, pgid)
		})
	}
}

// TestMagex_CustomTargetHandlesSIGTERM proves that a SIGTERM sent to magex
// alone, as a process supervisor or a CI runner may send it, reaches the
// custom target as an interrupt, which mage would not pass on: the target
// finishes its own shutdown, magex exits 143 without reporting a failure, and
// it takes about the target's cleanup time, not the 10 seconds before a kill.
func TestMagex_CustomTargetHandlesSIGTERM(t *testing.T) {
	if !hasMageBinary() {
		t.Skip("Skipping: mage binary not installed (required for custom magefile execution)")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("Skipping: go not installed (mage needs it to compile the magefile)")
	}
	if _, err := exec.LookPath("ps"); err != nil {
		t.Skip("Skipping: ps not installed (magex needs it to reach the compiled magefile)")
	}
	bin := buildMagexForSignalTest(t)

	cmd, out := startMagexServe(t, bin, serveMagefile)
	pgid := cmd.Process.Pid

	signaled := time.Now()
	require.NoError(t, syscall.Kill(cmd.Process.Pid, syscall.SIGTERM))

	exitCode, timedOut := waitOrTimeout(t, cmd, 30*time.Second)
	exited := time.Since(signaled)
	t.Logf("magex exited with %d at %+.2fs; output, timed from the SIGTERM:\n%s", exitCode, exited.Seconds(), out.Since(signaled))

	require.False(t, timedOut, "magex did not exit within 30s of SIGTERM")
	assert.Equal(t, 143, exitCode)
	output := out.String()
	assert.Contains(t, output, "serve: interrupt received, shutting down...")
	assert.Contains(t, output, "serve: shutdown complete")
	assert.NotContains(t, output, "custom command failed")
	assert.Less(t, exited, 8*time.Second, "the target should stop on its own, well before magex would kill it")
	assertGroupGone(t, pgid)
}

// startMagexServe writes magefile to a new module, starts magex serve there in
// its own process group, like a shell job, and returns once the Serve target
// is running. Signaling the group reaches magex, mage, and the compiled
// magefile, as Ctrl-C does.
func startMagexServe(t *testing.T, bin, magefile string) (*exec.Cmd, *lineRecorder) {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/sigrepro\n\ngo 1.24\n"), secureFilePerm))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "magefile.go"), []byte(magefile), secureFilePerm))

	out := &lineRecorder{}
	cmd := testCommand(t, bin, "serve")
	cmd.Dir = dir
	cmd.Stdout = out
	cmd.Stderr = out
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	// Don't hang on output held open by anything left behind
	cmd.WaitDelay = 5 * time.Second

	require.NoError(t, cmd.Start())
	pgid := cmd.Process.Pid
	t.Cleanup(func() {
		// Best effort: if the test failed early, stop whatever is left
		_ = syscall.Kill(-pgid, syscall.SIGKILL) //nolint:errcheck // the group is normally gone already
	})

	// mage compiles the magefile first, which can take a while on a cold cache
	require.Eventually(t, func() bool {
		return strings.Contains(out.String(), "serve: running")
	}, 2*time.Minute, 50*time.Millisecond, "the Serve target never started; output:\n%s", out)
	return cmd, out
}

// assertGroupGone checks that nothing is left running in process group pgid.
func assertGroupGone(t *testing.T, pgid int) {
	t.Helper()
	assert.Eventually(t, func() bool {
		return errors.Is(syscall.Kill(-pgid, 0), syscall.ESRCH)
	}, 2*time.Second, 50*time.Millisecond, "processes were left running in magex's process group")
}
