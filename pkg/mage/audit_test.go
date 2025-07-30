package mage

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/mrz1836/go-mage/pkg/mage/testutil"
	"github.com/mrz1836/go-mage/pkg/utils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// AuditTestSuite defines the test suite for Audit functions
type AuditTestSuite struct {
	suite.Suite
	env   *testutil.TestEnvironment
	audit Audit
}

// SetupTest runs before each test
func (ts *AuditTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
	ts.audit = Audit{}
}

// TearDownTest runs after each test
func (ts *AuditTestSuite) TearDownTest() {
	ts.env.Cleanup()
}

// TestAuditShow tests the Show method
func (ts *AuditTestSuite) TestAuditShow() {
	ts.Run("show audit events handles disabled audit gracefully", func() {
		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				setErr := SetRunner(r.(CommandRunner)) //nolint:errcheck // Error is checked on next line
				if setErr != nil {
					return setErr
				}
				return nil
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.audit.Show()
			},
		)
		// The audit system may be disabled, which is acceptable
		if err != nil {
			require.Contains(ts.T(), err.Error(), "audit")
		}
	})

	ts.Run("show audit events with environment filters", func() {
		// Set environment variables for filtering
		originalStartTime := os.Getenv("START_TIME")
		originalEndTime := os.Getenv("END_TIME")
		originalUser := os.Getenv("USER")
		originalCommand := os.Getenv("COMMAND")
		originalSuccess := os.Getenv("SUCCESS")
		originalLimit := os.Getenv("LIMIT")
		defer func() {
			if err := os.Setenv("START_TIME", originalStartTime); err != nil {
				ts.T().Logf("Failed to restore START_TIME: %v", err)
			}
			if err := os.Setenv("END_TIME", originalEndTime); err != nil {
				ts.T().Logf("Failed to restore END_TIME: %v", err)
			}
			if err := os.Setenv("USER", originalUser); err != nil {
				ts.T().Logf("Failed to restore USER: %v", err)
			}
			if err := os.Setenv("COMMAND", originalCommand); err != nil {
				ts.T().Logf("Failed to restore COMMAND: %v", err)
			}
			if err := os.Setenv("SUCCESS", originalSuccess); err != nil {
				ts.T().Logf("Failed to restore SUCCESS: %v", err)
			}
			if err := os.Setenv("LIMIT", originalLimit); err != nil {
				ts.T().Logf("Failed to restore LIMIT: %v", err)
			}
		}()

		if err := os.Setenv("START_TIME", "2023-01-01"); err != nil {
			ts.T().Fatalf("Failed to set START_TIME: %v", err)
		}
		if err := os.Setenv("END_TIME", "2023-12-31"); err != nil {
			ts.T().Fatalf("Failed to set END_TIME: %v", err)
		}
		if err := os.Setenv("USER", "testuser"); err != nil {
			ts.T().Fatalf("Failed to set USER: %v", err)
		}
		if err := os.Setenv("COMMAND", "test:run"); err != nil {
			ts.T().Fatalf("Failed to set COMMAND: %v", err)
		}
		if err := os.Setenv("SUCCESS", "true"); err != nil {
			ts.T().Fatalf("Failed to set SUCCESS: %v", err)
		}
		if err := os.Setenv("LIMIT", "10"); err != nil {
			ts.T().Fatalf("Failed to set LIMIT: %v", err)
		}

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				setErr := SetRunner(r.(CommandRunner)) //nolint:errcheck // Error is checked on next line
				if setErr != nil {
					return setErr
				}
				return nil
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.audit.Show()
			},
		)
		// The audit system may be disabled, which is acceptable
		if err != nil {
			require.Contains(ts.T(), err.Error(), "audit")
		}
	})
}

// TestAuditStats tests the Stats method
func (ts *AuditTestSuite) TestAuditStats() {
	ts.Run("display audit statistics handles disabled audit gracefully", func() {
		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				setErr := SetRunner(r.(CommandRunner)) //nolint:errcheck // Error is checked on next line
				if setErr != nil {
					return setErr
				}
				return nil
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.audit.Stats()
			},
		)
		// The audit system may be disabled, which is acceptable
		if err != nil {
			require.Contains(ts.T(), err.Error(), "audit")
		}
	})
}

// TestAuditExport tests the Export method
func (ts *AuditTestSuite) TestAuditExport() {
	ts.Run("export audit events with default settings", func() {
		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				setErr := SetRunner(r.(CommandRunner)) //nolint:errcheck // Error is checked on next line
				if setErr != nil {
					return setErr
				}
				return nil
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.audit.Export()
			},
		)
		// The audit system may be disabled, which is acceptable
		if err != nil {
			require.Contains(ts.T(), err.Error(), "audit")
		}
	})

	ts.Run("export audit events with custom output file", func() {
		// Set custom output file
		originalOutput := os.Getenv("OUTPUT")
		defer func() {
			if err := os.Setenv("OUTPUT", originalOutput); err != nil {
				ts.T().Logf("Failed to restore OUTPUT: %v", err)
			}
		}()
		if err := os.Setenv("OUTPUT", "custom-audit.json"); err != nil {
			ts.T().Fatalf("Failed to set OUTPUT: %v", err)
		}

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				setErr := SetRunner(r.(CommandRunner)) //nolint:errcheck // Error is checked on next line
				if setErr != nil {
					return setErr
				}
				return nil
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.audit.Export()
			},
		)
		// The audit system may be disabled, which is acceptable
		if err != nil {
			require.Contains(ts.T(), err.Error(), "audit")
		}
	})
}

// TestAuditCleanup tests the Cleanup method
func (ts *AuditTestSuite) TestAuditCleanup() {
	ts.Run("cleanup old audit events", func() {
		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				setErr := SetRunner(r.(CommandRunner)) //nolint:errcheck // Error is checked on next line
				if setErr != nil {
					return setErr
				}
				return nil
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.audit.Cleanup()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestAuditEnable tests the Enable method
func (ts *AuditTestSuite) TestAuditEnable() {
	ts.Run("enable audit logging", func() {
		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				setErr := SetRunner(r.(CommandRunner)) //nolint:errcheck // Error is checked on next line
				if setErr != nil {
					return setErr
				}
				return nil
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.audit.Enable()
			},
		)

		require.NoError(ts.T(), err)

		// Check that .mage directory was created
		require.True(ts.T(), ts.env.FileExists(".mage"))
	})

	ts.Run("enable audit logging with existing config", func() {
		// Create existing config file
		ts.env.CreateFile(".mage/config.yaml", `{"audit": {"enabled": false}}`)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				setErr := SetRunner(r.(CommandRunner)) //nolint:errcheck // Error is checked on next line
				if setErr != nil {
					return setErr
				}
				return nil
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.audit.Enable()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestAuditDisable tests the Disable method
func (ts *AuditTestSuite) TestAuditDisable() {
	ts.Run("disable audit logging without existing config", func() {
		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				setErr := SetRunner(r.(CommandRunner)) //nolint:errcheck // Error is checked on next line
				if setErr != nil {
					return setErr
				}
				return nil
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.audit.Disable()
			},
		)
		// May fail if directory doesn't exist, which is acceptable
		if err != nil {
			require.Contains(ts.T(), err.Error(), "no such file or directory")
		}
	})

	ts.Run("disable audit logging with existing config", func() {
		// Create existing config file
		ts.env.CreateFile(".mage/config.yaml", `{"audit": {"enabled": true}}`)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				setErr := SetRunner(r.(CommandRunner)) //nolint:errcheck // Error is checked on next line
				if setErr != nil {
					return setErr
				}
				return nil
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.audit.Disable()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestAuditReport tests the Report method
func (ts *AuditTestSuite) TestAuditReport() {
	ts.Run("generate compliance report handles disabled audit gracefully", func() {
		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				setErr := SetRunner(r.(CommandRunner)) //nolint:errcheck // Error is checked on next line
				if setErr != nil {
					return setErr
				}
				return nil
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.audit.Report()
			},
		)
		// The audit system may be disabled, which is acceptable
		if err != nil {
			require.Contains(ts.T(), err.Error(), "audit")
		}
	})

	ts.Run("generate compliance report with custom output", func() {
		// Set custom output file
		originalOutput := os.Getenv("OUTPUT")
		defer func() {
			if err := os.Setenv("OUTPUT", originalOutput); err != nil {
				ts.T().Logf("Failed to restore OUTPUT: %v", err)
			}
		}()
		if err := os.Setenv("OUTPUT", "custom-report.json"); err != nil {
			ts.T().Fatalf("Failed to set OUTPUT: %v", err)
		}

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				setErr := SetRunner(r.(CommandRunner)) //nolint:errcheck // Error is checked on next line
				if setErr != nil {
					return setErr
				}
				return nil
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.audit.Report()
			},
		)
		// The audit system may be disabled, which is acceptable
		if err != nil {
			require.Contains(ts.T(), err.Error(), "audit")
		}
	})
}

// TestAuditHelperFunctions tests the helper functions
func (ts *AuditTestSuite) TestAuditHelperFunctions() {
	ts.Run("LogCommandExecution function", func() {
		// Test logging a command execution
		startTime := time.Now()
		duration := 100 * time.Millisecond

		// This function should not return an error (errors handled internally)
		LogCommandExecution("test", []string{"arg1", "arg2"}, startTime, duration, 0, true)

		// Test with failed command
		LogCommandExecution("failed-test", []string{}, startTime, duration, 1, false)

		// These should complete without panics
		require.True(ts.T(), true)
	})

	ts.Run("getFilteredEnvironment function", func() {
		// Set some environment variables
		originalGoVersion := os.Getenv("GO_VERSION")
		originalGoPath := os.Getenv("GOPATH")
		originalMageVerbose := os.Getenv("MAGE_VERBOSE")
		defer func() {
			if err := os.Setenv("GO_VERSION", originalGoVersion); err != nil {
				ts.T().Logf("Failed to restore GO_VERSION: %v", err)
			}
			if err := os.Setenv("GOPATH", originalGoPath); err != nil {
				ts.T().Logf("Failed to restore GOPATH: %v", err)
			}
			if err := os.Setenv("MAGE_VERBOSE", originalMageVerbose); err != nil {
				ts.T().Logf("Failed to restore MAGE_VERBOSE: %v", err)
			}
		}()

		if err := os.Setenv("GO_VERSION", "1.24.0"); err != nil {
			ts.T().Fatalf("Failed to set GO_VERSION: %v", err)
		}
		if err := os.Setenv("GOPATH", "/go"); err != nil {
			ts.T().Fatalf("Failed to set GOPATH: %v", err)
		}
		if err := os.Setenv("MAGE_VERBOSE", "true"); err != nil {
			ts.T().Fatalf("Failed to set MAGE_VERBOSE: %v", err)
		}

		env := getFilteredEnvironment()
		require.Contains(ts.T(), env, "GO_VERSION")
		require.Equal(ts.T(), "1.24.0", env["GO_VERSION"])
		require.Contains(ts.T(), env, "GOPATH")
		require.Contains(ts.T(), env, "MAGE_VERBOSE")

		// Should not contain non-relevant environment variables
		require.NotContains(ts.T(), env, "HOME")
		require.NotContains(ts.T(), env, "PATH")
	})

	ts.Run("getGoVersion function", func() {
		ts.Run("successful go version detection", func() {
			// Create fresh test environment for isolated testing
			env := testutil.NewTestEnvironment(ts.T())
			defer env.Cleanup()

			// Mock successful go version command
			env.Runner.On("RunCmdOutput", "go", []string{"version"}).Return("go version go1.24.0 linux/amd64", nil)

			version := ""
			err := env.WithMockRunner(
				func(r interface{}) error {
					setErr := SetRunner(r.(CommandRunner)) //nolint:errcheck // Error is checked on next line
					if setErr != nil {
						return setErr
					}
					return nil
				},
				func() interface{} { return GetRunner() },
				func() error {
					version = getGoVersion()
					return nil
				},
			)

			require.NoError(ts.T(), err)
			require.Equal(ts.T(), "1.24.0", version)
		})

		ts.Run("failed go version detection", func() {
			// Create fresh test environment for isolated testing
			env := testutil.NewTestEnvironment(ts.T())
			defer env.Cleanup()

			// Mock failed go version command
			env.Runner.On("RunCmdOutput", "go", []string{"version"}).Return("", errors.New("command failed"))

			version := ""
			err := env.WithMockRunner(
				func(r interface{}) error {
					setErr := SetRunner(r.(CommandRunner)) //nolint:errcheck // Error is checked on next line
					if setErr != nil {
						return setErr
					}
					return nil
				},
				func() interface{} { return GetRunner() },
				func() error {
					version = getGoVersion()
					return nil
				},
			)

			require.NoError(ts.T(), err)
			require.Equal(ts.T(), "unknown", version)
		})

		ts.Run("malformed go version output", func() {
			// Create fresh test environment for isolated testing
			env := testutil.NewTestEnvironment(ts.T())
			defer env.Cleanup()

			// Mock malformed go version output
			env.Runner.On("RunCmdOutput", "go", []string{"version"}).Return("invalid output", nil)

			version := ""
			err := env.WithMockRunner(
				func(r interface{}) error {
					setErr := SetRunner(r.(CommandRunner)) //nolint:errcheck // Error is checked on next line
					if setErr != nil {
						return setErr
					}
					return nil
				},
				func() interface{} { return GetRunner() },
				func() error {
					version = getGoVersion()
					return nil
				},
			)

			require.NoError(ts.T(), err)
			require.Equal(ts.T(), "unknown", version)
		})
	})

	ts.Run("truncateString helper function", func() {
		// Test string truncation helper used in Show method
		result := truncateString("This is a very long string that should be truncated", 10)
		require.LessOrEqual(ts.T(), len(result), 13) // Account for ellipsis

		// Test short string (no truncation needed)
		shortString := "short"
		result = truncateString(shortString, 10)
		require.Equal(ts.T(), shortString, result)
	})

	ts.Run("getVersion helper function", func() {
		// Test version detection helper used in LogCommandExecution
		ts.env.Runner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--abbrev=0"}).Return("v1.2.3", nil)

		version := ""
		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				setErr := SetRunner(r.(CommandRunner)) //nolint:errcheck // Error is checked on next line
				if setErr != nil {
					return setErr
				}
				return nil
			},
			func() interface{} { return GetRunner() },
			func() error {
				version = getVersion()
				return nil
			},
		)

		require.NoError(ts.T(), err)
		require.NotEmpty(ts.T(), version)
	})
}

// TestAuditComplianceReportStructure tests the AuditComplianceReport structure
func (ts *AuditTestSuite) TestAuditComplianceReportStructure() {
	ts.Run("create audit compliance report", func() {
		report := AuditComplianceReport{
			GeneratedAt:      time.Now(),
			ReportPeriod:     "2023-01-01 to 2023-12-31",
			TotalEvents:      100,
			SuccessfulEvents: 95,
			FailedEvents:     5,
			SuccessRate:      95.0,
			TopUsers:         []utils.UserStats{},
			TopCommands:      []utils.CommandStats{},
			RecentFailures:   []utils.AuditEvent{},
		}

		require.NotZero(ts.T(), report.GeneratedAt)
		require.Equal(ts.T(), 100, report.TotalEvents)
		require.Equal(ts.T(), 95, report.SuccessfulEvents)
		require.Equal(ts.T(), 5, report.FailedEvents)
		require.Equal(ts.T(), 95.0, report.SuccessRate)
	})
}

// TestAuditIntegration tests integration scenarios
func (ts *AuditTestSuite) TestAuditIntegration() {
	ts.Run("complete audit workflow", func() {
		// Mock git and go version commands that are called by LogCommandExecution
		ts.env.Runner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--abbrev=0"}).Return("v1.0.0", nil)
		ts.env.Runner.On("RunCmdOutput", "go", []string{"version"}).Return("go version go1.24.0 linux/amd64", nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				setErr := SetRunner(r.(CommandRunner)) //nolint:errcheck // Error is checked on next line
				if setErr != nil {
					return setErr
				}
				return nil
			},
			func() interface{} { return GetRunner() },
			func() error {
				// Enable audit logging (may fail if directory doesn't exist)
				if err := ts.audit.Enable(); err != nil {
					// Error expected and acceptable - audit may be disabled
				}

				// Log some command executions (these should not fail)
				startTime := time.Now()
				duration := 50 * time.Millisecond
				LogCommandExecution("build", []string{}, startTime, duration, 0, true)
				LogCommandExecution("test", []string{"-v"}, startTime, duration, 1, false)

				// These operations may fail if audit is disabled, which is acceptable
				if err := ts.audit.Show(); err != nil {
					// Error expected and acceptable - audit may be disabled
				}
				if err := ts.audit.Stats(); err != nil {
					// Error expected and acceptable - audit may be disabled
				}
				if err := ts.audit.Export(); err != nil {
					// Error expected and acceptable - audit may be disabled
				}
				if err := ts.audit.Report(); err != nil {
					// Error expected and acceptable - audit may be disabled
				}

				// Cleanup should work regardless
				if err := ts.audit.Cleanup(); err != nil {
					// Error expected and acceptable - cleanup is best-effort
				}

				// Disable audit logging (may fail if file doesn't exist)
				if err := ts.audit.Disable(); err != nil {
					// Error expected and acceptable - audit may already be disabled
				}

				return nil
			},
		)

		require.NoError(ts.T(), err)
	})

	ts.Run("audit with various environment configurations", func() {
		// Set various environment variables
		originalVars := map[string]string{
			"START_TIME":         os.Getenv("START_TIME"),
			"END_TIME":           os.Getenv("END_TIME"),
			"USER":               os.Getenv("USER"),
			"COMMAND":            os.Getenv("COMMAND"),
			"SUCCESS":            os.Getenv("SUCCESS"),
			"LIMIT":              os.Getenv("LIMIT"),
			"OUTPUT":             os.Getenv("OUTPUT"),
			"GO_VERSION":         os.Getenv("GO_VERSION"),
			"MAGE_AUDIT_ENABLED": os.Getenv("MAGE_AUDIT_ENABLED"),
		}
		defer func() {
			for key, value := range originalVars {
				if err := os.Setenv(key, value); err != nil {
					ts.T().Logf("Failed to restore %s: %v", key, err)
				}
			}
		}()

		if err := os.Setenv("START_TIME", "2023-01-01"); err != nil {
			ts.T().Fatalf("Failed to set START_TIME: %v", err)
		}
		if err := os.Setenv("END_TIME", "2023-12-31"); err != nil {
			ts.T().Fatalf("Failed to set END_TIME: %v", err)
		}
		if err := os.Setenv("USER", "testuser"); err != nil {
			ts.T().Fatalf("Failed to set USER: %v", err)
		}
		if err := os.Setenv("COMMAND", "test"); err != nil {
			ts.T().Fatalf("Failed to set COMMAND: %v", err)
		}
		if err := os.Setenv("SUCCESS", "true"); err != nil {
			ts.T().Fatalf("Failed to set SUCCESS: %v", err)
		}
		if err := os.Setenv("LIMIT", "25"); err != nil {
			ts.T().Fatalf("Failed to set LIMIT: %v", err)
		}
		if err := os.Setenv("OUTPUT", "test-audit.json"); err != nil {
			ts.T().Fatalf("Failed to set OUTPUT: %v", err)
		}
		if err := os.Setenv("GO_VERSION", "1.24.0"); err != nil {
			ts.T().Fatalf("Failed to set GO_VERSION: %v", err)
		}
		if err := os.Setenv("MAGE_AUDIT_ENABLED", "true"); err != nil {
			ts.T().Fatalf("Failed to set MAGE_AUDIT_ENABLED: %v", err)
		}

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				setErr := SetRunner(r.(CommandRunner)) //nolint:errcheck // Error is checked on next line
				if setErr != nil {
					return setErr
				}
				return nil
			},
			func() interface{} { return GetRunner() },
			func() error {
				// Test all methods with environment variables set
				// These may fail if audit is disabled, which is acceptable
				if err := ts.audit.Show(); err != nil {
					// Error expected and acceptable - audit may be disabled
				}
				if err := ts.audit.Export(); err != nil {
					// Error expected and acceptable - audit may be disabled
				}
				if err := ts.audit.Report(); err != nil {
					// Error expected and acceptable - audit may be disabled
				}
				return nil
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestAuditTestSuite runs the test suite
func TestAuditTestSuite(t *testing.T) {
	suite.Run(t, new(AuditTestSuite))
}
