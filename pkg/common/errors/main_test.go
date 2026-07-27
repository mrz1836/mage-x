package errors

import (
	"fmt"
	"os"
	"testing"

	"github.com/mrz1836/mage-x/pkg/testhelpers"
)

// TestMain keeps the package hermetic. The webhook notifier posts over HTTP, so
// block anything that is not loopback; httptest servers keep working.
func TestMain(m *testing.M) {
	if err := testhelpers.BlockExternalNetwork(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to block external network: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}
