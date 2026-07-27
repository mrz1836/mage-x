package utils

import (
	"fmt"
	"os"
	"testing"

	"github.com/mrz1836/mage-x/pkg/testhelpers"
)

// TestMain keeps the package hermetic. utils owns the shared HTTP client and the
// download helpers, so a mistake here is the easiest way for a test run to reach
// the internet; loopback stays available for httptest servers.
//
// The client's transport is left alone (tests here assert its configuration);
// it honors the proxy variables set below, which is what keeps it local.
func TestMain(m *testing.M) {
	if err := testhelpers.BlockExternalNetwork(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to block external network: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}
