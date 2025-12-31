package mage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// OperationsCoverageTestSuite provides comprehensive coverage for operations.go types
type OperationsCoverageTestSuite struct {
	suite.Suite

	env *testutil.TestEnvironment
}

func TestOperationsCoverageTestSuite(t *testing.T) {
	suite.Run(t, new(OperationsCoverageTestSuite))
}

func (ts *OperationsCoverageTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")

	// Set up general mocks for all commands - catch-all approach
	ts.env.Runner.On("RunCmd", mock.Anything, mock.Anything).Return(nil).Maybe()
	ts.env.Runner.On("RunCmdOutput", mock.Anything, mock.Anything).Return("", nil).Maybe()
	ts.env.Runner.On("RunCmdOutputDir", mock.Anything, mock.Anything, mock.Anything).Return("", nil).Maybe()
	ts.env.Runner.On("RunCmdInDir", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
}

func (ts *OperationsCoverageTestSuite) TearDownTest() {
	TestResetConfig()
	ts.env.Cleanup()
}

// Helper to set up mock runner and execute function
func (ts *OperationsCoverageTestSuite) withMockRunner(fn func() error) error {
	return ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		fn,
	)
}

// ================== Check Namespace Tests ==================

func (ts *OperationsCoverageTestSuite) TestCheckAllSuccess() {
	err := ts.withMockRunner(func() error {
		return Check{}.All()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCheckFormatSuccess() {
	err := ts.withMockRunner(func() error {
		return Check{}.Format()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCheckImportsSuccess() {
	err := ts.withMockRunner(func() error {
		return Check{}.Imports()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCheckLicenseSuccess() {
	err := ts.withMockRunner(func() error {
		return Check{}.License()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCheckSecuritySuccess() {
	err := ts.withMockRunner(func() error {
		return Check{}.Security()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCheckDependenciesSuccess() {
	err := ts.withMockRunner(func() error {
		return Check{}.Dependencies()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCheckTidySuccess() {
	err := ts.withMockRunner(func() error {
		return Check{}.Tidy()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCheckGenerateSuccess() {
	err := ts.withMockRunner(func() error {
		return Check{}.Generate()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCheckSpellingSuccess() {
	err := ts.withMockRunner(func() error {
		return Check{}.Spelling()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCheckDocumentationSuccess() {
	err := ts.withMockRunner(func() error {
		return Check{}.Documentation()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCheckDepsSuccess() {
	err := ts.withMockRunner(func() error {
		return Check{}.Deps()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCheckDocsSuccess() {
	err := ts.withMockRunner(func() error {
		return Check{}.Docs()
	})
	ts.Require().NoError(err)
}

// ================== CI Namespace Tests ==================

func (ts *OperationsCoverageTestSuite) TestCISetupGitHub() {
	err := ts.withMockRunner(func() error {
		return CI{}.Setup("github")
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCISetupGitLab() {
	err := ts.withMockRunner(func() error {
		return CI{}.Setup("gitlab")
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCISetupJenkins() {
	err := ts.withMockRunner(func() error {
		return CI{}.Setup("jenkins")
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCISetupCircleCI() {
	err := ts.withMockRunner(func() error {
		return CI{}.Setup("circleci")
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCISetupUnsupported() {
	err := ts.withMockRunner(func() error {
		return CI{}.Setup("unsupported")
	})
	ts.Require().Error(err)
	ts.Contains(err.Error(), "unsupported")
}

func (ts *OperationsCoverageTestSuite) TestCIValidateSuccess() {
	err := ts.withMockRunner(func() error {
		return CI{}.Validate()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCIRunWithJob() {
	err := ts.withMockRunner(func() error {
		return CI{}.Run("test")
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCIRunWithoutJob() {
	err := ts.withMockRunner(func() error {
		return CI{}.Run("")
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCIStatusSuccess() {
	err := ts.withMockRunner(func() error {
		return CI{}.Status("main")
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCILogsSuccess() {
	err := ts.withMockRunner(func() error {
		return CI{}.Logs("12345")
	})
	ts.Require().NoError(err)
}

// ================== Monitor Namespace Tests ==================

func (ts *OperationsCoverageTestSuite) TestMonitorStartSuccess() {
	err := ts.withMockRunner(func() error {
		return Monitor{}.Start()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestMonitorStopSuccess() {
	err := ts.withMockRunner(func() error {
		return Monitor{}.Stop()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestMonitorStatusSuccess() {
	err := ts.withMockRunner(func() error {
		return Monitor{}.Status()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestMonitorLogsSuccess() {
	err := ts.withMockRunner(func() error {
		return Monitor{}.Logs("web")
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestMonitorMetricsSuccess() {
	err := ts.withMockRunner(func() error {
		return Monitor{}.Metrics("1h")
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestMonitorAlertsSuccess() {
	err := ts.withMockRunner(func() error {
		return Monitor{}.Alerts()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestMonitorHealthSuccess() {
	err := ts.withMockRunner(func() error {
		return Monitor{}.Health("/health")
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestMonitorDashboardSuccess() {
	err := ts.withMockRunner(func() error {
		return Monitor{}.Dashboard(8080)
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestMonitorTraceSuccess() {
	err := ts.withMockRunner(func() error {
		return Monitor{}.Trace("trace-123")
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestMonitorProfileSuccess() {
	err := ts.withMockRunner(func() error {
		return Monitor{}.Profile("cpu", "30s")
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestMonitorExportSuccess() {
	err := ts.withMockRunner(func() error {
		return Monitor{}.Export("json", "1h")
	})
	ts.Require().NoError(err)
}

// ================== Database Namespace Tests ==================

func (ts *OperationsCoverageTestSuite) TestDatabaseMigrateSuccess() {
	err := ts.withMockRunner(func() error {
		return Database{}.Migrate("up")
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestDatabaseSeedSuccess() {
	err := ts.withMockRunner(func() error {
		return Database{}.Seed("test-data")
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestDatabaseResetSuccess() {
	err := ts.withMockRunner(func() error {
		return Database{}.Reset()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestDatabaseBackupSuccess() {
	err := ts.withMockRunner(func() error {
		return Database{}.Backup("backup.sql")
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestDatabaseRestoreSuccess() {
	err := ts.withMockRunner(func() error {
		return Database{}.Restore("backup.sql")
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestDatabaseStatusSuccess() {
	err := ts.withMockRunner(func() error {
		return Database{}.Status()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestDatabaseCreateSuccess() {
	err := ts.withMockRunner(func() error {
		return Database{}.Create("testdb")
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestDatabaseDropSuccess() {
	err := ts.withMockRunner(func() error {
		return Database{}.Drop("testdb")
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestDatabaseConsoleSuccess() {
	err := ts.withMockRunner(func() error {
		return Database{}.Console()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestDatabaseQuerySuccess() {
	err := ts.withMockRunner(func() error {
		return Database{}.Query("SELECT 1")
	})
	ts.Require().NoError(err)
}

// ================== Deploy Namespace Tests ==================

func (ts *OperationsCoverageTestSuite) TestDeployLocalSuccess() {
	err := ts.withMockRunner(func() error {
		return Deploy{}.Local()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestDeployProductionSuccess() {
	err := ts.withMockRunner(func() error {
		return Deploy{}.Production()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestDeployStagingSuccess() {
	err := ts.withMockRunner(func() error {
		return Deploy{}.Staging()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestDeployKubernetesSuccess() {
	err := ts.withMockRunner(func() error {
		return Deploy{}.Kubernetes("default")
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestDeployAWSSuccess() {
	err := ts.withMockRunner(func() error {
		return Deploy{}.AWS("lambda")
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestDeployGCPSuccess() {
	err := ts.withMockRunner(func() error {
		return Deploy{}.GCP("cloud-run")
	})
	ts.Require().NoError(err)
}

// ================== Clean Namespace Tests ==================

func (ts *OperationsCoverageTestSuite) TestCleanAllSuccess() {
	err := ts.withMockRunner(func() error {
		return Clean{}.All()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCleanBuildSuccess() {
	err := ts.withMockRunner(func() error {
		return Clean{}.Build()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCleanTestSuccess() {
	err := ts.withMockRunner(func() error {
		return Clean{}.Test()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCleanCacheSuccess() {
	err := ts.withMockRunner(func() error {
		return Clean{}.Cache()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCleanDependenciesSuccess() {
	err := ts.withMockRunner(func() error {
		return Clean{}.Dependencies()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCleanDepsSuccess() {
	err := ts.withMockRunner(func() error {
		return Clean{}.Deps()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCleanFullSuccess() {
	err := ts.withMockRunner(func() error {
		return Clean{}.Full()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCleanGeneratedSuccess() {
	err := ts.withMockRunner(func() error {
		return Clean{}.Generated()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCleanDistSuccess() {
	err := ts.withMockRunner(func() error {
		return Clean{}.Dist()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCleanLogsSuccess() {
	err := ts.withMockRunner(func() error {
		return Clean{}.Logs()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestCleanTempSuccess() {
	err := ts.withMockRunner(func() error {
		return Clean{}.Temp()
	})
	ts.Require().NoError(err)
}

// ================== Run Namespace Tests ==================

func (ts *OperationsCoverageTestSuite) TestRunDevSuccess() {
	err := ts.withMockRunner(func() error {
		return Run{}.Dev()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestRunProdSuccess() {
	err := ts.withMockRunner(func() error {
		return Run{}.Prod()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestRunWatchSuccess() {
	err := ts.withMockRunner(func() error {
		return Run{}.Watch()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestRunDebugSuccess() {
	err := ts.withMockRunner(func() error {
		return Run{}.Debug()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestRunProfileSuccess() {
	err := ts.withMockRunner(func() error {
		return Run{}.Profile()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestRunBenchmarkSuccess() {
	err := ts.withMockRunner(func() error {
		return Run{}.Benchmark()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestRunServerSuccess() {
	err := ts.withMockRunner(func() error {
		return Run{}.Server()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestRunMigrationsSuccess() {
	err := ts.withMockRunner(func() error {
		return Run{}.Migrations()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestRunSeedsSuccess() {
	err := ts.withMockRunner(func() error {
		return Run{}.Seeds()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestRunWorkerSuccess() {
	err := ts.withMockRunner(func() error {
		return Run{}.Worker()
	})
	ts.Require().NoError(err)
}

// ================== Serve Namespace Tests ==================

func (ts *OperationsCoverageTestSuite) TestServeHTTPSuccess() {
	err := ts.withMockRunner(func() error {
		return Serve{}.HTTP()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestServeHTTPSSuccess() {
	err := ts.withMockRunner(func() error {
		return Serve{}.HTTPS()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestServeDocsSuccess() {
	err := ts.withMockRunner(func() error {
		return Serve{}.Docs()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestServeAPISuccess() {
	err := ts.withMockRunner(func() error {
		return Serve{}.API()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestServeGRPCSuccess() {
	err := ts.withMockRunner(func() error {
		return Serve{}.GRPC()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestServeMetricsSuccess() {
	err := ts.withMockRunner(func() error {
		return Serve{}.Metrics()
	})
	ts.Require().NoError(err)
}

func (ts *OperationsCoverageTestSuite) TestServeStaticSuccess() {
	err := ts.withMockRunner(func() error {
		return Serve{}.Static()
	})
	ts.Require().NoError(err)
}

// ================== Standalone Tests ==================

func TestOperationsStaticErrors(t *testing.T) {
	assert.Error(t, errUnsupportedCIProvider)
}
