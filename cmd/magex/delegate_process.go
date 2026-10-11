package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// processListTimeout bounds the ps call that lists the running processes.
const processListTimeout = 2 * time.Second

// errProcessListUnsupported is returned where magex can't list processes.
var errProcessListUnsupported = errors.New("listing processes is not supported on this platform")

// processChildren returns the running processes' PIDs, keyed by their parent's
// PID. magex starts mage, and mage starts the compiled magefile, so this is how
// magex reaches the process that runs the target. It runs ps, which macOS,
// Linux, and the BSDs all have with these options, rather than reading each
// platform's process table its own way.
func processChildren() (map[int][]int, error) {
	if runtime.GOOS == "windows" {
		return nil, errProcessListUnsupported
	}
	// Not the delegation's context: it is canceled by the time this runs
	ctx, cancel := context.WithTimeout(context.Background(), processListTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "ps", "-A", "-o", "pid=", "-o", "ppid=").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %w", err)
	}
	return parseProcessList(string(out)), nil
}

// parseProcessList reads ps output, one "pid ppid" pair per line, into PIDs
// keyed by their parent's PID. Lines it can't read are skipped.
func parseProcessList(out string) map[int][]int {
	children := make(map[int][]int)
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		pid, pidErr := strconv.Atoi(fields[0])
		ppid, ppidErr := strconv.Atoi(fields[1])
		if pidErr != nil || ppidErr != nil || pid == ppid {
			continue
		}
		children[ppid] = append(children[ppid], pid)
	}
	return children
}

// descendantPIDs returns every process below pid in children: its children,
// their children, and so on, parents before their children.
func descendantPIDs(children map[int][]int, pid int) []int {
	var found []int
	seen := map[int]bool{pid: true}
	queue := []int{pid}
	for len(queue) > 0 {
		parent := queue[0]
		queue = queue[1:]
		for _, child := range children[parent] {
			if seen[child] {
				continue
			}
			seen[child] = true
			found = append(found, child)
			queue = append(queue, child)
		}
	}
	return found
}

// signalPIDs sends sig to each of pids, as a best effort: a process may have
// exited since it was listed.
func signalPIDs(pids []int, sig os.Signal) {
	for _, pid := range pids {
		proc, err := os.FindProcess(pid)
		if err != nil {
			continue
		}
		_ = proc.Signal(sig) //nolint:errcheck // best effort: the process may have exited since it was listed
		_ = proc.Release()   //nolint:errcheck // nothing to do if releasing the handle fails
	}
}
