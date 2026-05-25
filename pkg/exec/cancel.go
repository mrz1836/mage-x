package exec

import (
	"os"
	"os/exec"
	"time"
)

// gracefulCancelDelay is the window between SIGINT delivery and the runtime's
// follow-up SIGKILL. Long enough for golangci-lint / go test to flush partial
// output and clean up tempdirs; short enough that a hung child never blocks
// the user for more than this duration.
const gracefulCancelDelay = 5 * time.Second

// applyGracefulCancel wires cmd.Cancel + cmd.WaitDelay so that when the
// command's context is canceled the child receives os.Interrupt first
// (graceful shutdown) and is escalated to SIGKILL only after the grace
// window elapses. Works on Unix; on Windows os.Interrupt is best-effort but
// WaitDelay still guarantees termination.
//
// Requires Go 1.20+ (mage-x's go.mod is on 1.25).
func applyGracefulCancel(cmd *exec.Cmd) {
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		return cmd.Process.Signal(os.Interrupt)
	}
	cmd.WaitDelay = gracefulCancelDelay
}
