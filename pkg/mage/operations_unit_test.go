package mage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// operationsTestHelper provides test utilities for operations testing
type operationsTestHelper struct {
	originalRunner CommandRunner
	mockRunner     *testutil.MockRunner
}

// newOperationsTestHelper creates a new operations test helper
func newOperationsTestHelper(tb testing.TB) *operationsTestHelper {
	tb.Helper()
	h := &operationsTestHelper{
		originalRunner: GetRunner(),
	}
	runner, _ := testutil.NewMockRunner()
	h.mockRunner = runner
	require.NoError(tb, SetRunner(h.mockRunner))
	return h
}

// teardown restores the original runner
func (h *operationsTestHelper) teardown(tb testing.TB) {
	tb.Helper()
	if h.originalRunner != nil {
		require.NoError(tb, SetRunner(h.originalRunner))
	}
}

// expectCmd sets up an expectation for a command
func (h *operationsTestHelper) expectCmd(cmd string) {
	h.mockRunner.On("RunCmd", cmd, mock.Anything).Return(nil)
}

// expectCmdError sets up an expectation for a command that returns an error
func (h *operationsTestHelper) expectCmdError(cmd string) {
	h.mockRunner.On("RunCmd", cmd, mock.Anything).Return(assert.AnError)
}

// TestCheckAllSuccess tests Check.All success path
func TestCheckAllSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Check{}.All()
	require.NoError(t, err)
}

// TestCheckAllError tests Check.All error path
func TestCheckAllError(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmdError("echo")

	err := Check{}.All()
	require.Error(t, err)
}

// TestCheckFormatSuccess tests Check.Format success path
func TestCheckFormatSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("gofmt")

	err := Check{}.Format()
	require.NoError(t, err)
}

// TestCheckFormatError tests Check.Format error path
func TestCheckFormatError(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmdError("gofmt")

	err := Check{}.Format()
	require.Error(t, err)
}

// TestCheckImportsSuccess tests Check.Imports success path
func TestCheckImportsSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("goimports")

	err := Check{}.Imports()
	require.NoError(t, err)
}

// TestCheckImportsError tests Check.Imports error path
func TestCheckImportsError(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmdError("goimports")

	err := Check{}.Imports()
	require.Error(t, err)
}

// TestCheckLicenseSuccess tests Check.License success path
func TestCheckLicenseSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Check{}.License()
	require.NoError(t, err)
}

// TestCheckSecuritySuccess tests Check.Security success path
func TestCheckSecuritySuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("gosec")

	err := Check{}.Security()
	require.NoError(t, err)
}

// TestCheckSecurityError tests Check.Security error path
func TestCheckSecurityError(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmdError("gosec")

	err := Check{}.Security()
	require.Error(t, err)
}

// TestCheckDependenciesSuccess tests Check.Dependencies success path
func TestCheckDependenciesSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("go")

	err := Check{}.Dependencies()
	require.NoError(t, err)
}

// TestCheckDependenciesError tests Check.Dependencies error path
func TestCheckDependenciesError(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmdError("go")

	err := Check{}.Dependencies()
	require.Error(t, err)
}

// TestCheckTidySuccess tests Check.Tidy success path
func TestCheckTidySuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("go")

	err := Check{}.Tidy()
	require.NoError(t, err)
}

// TestCheckTidyError tests Check.Tidy error path
func TestCheckTidyError(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmdError("go")

	err := Check{}.Tidy()
	require.Error(t, err)
}

// TestCheckGenerateSuccess tests Check.Generate success path
func TestCheckGenerateSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("go")

	err := Check{}.Generate()
	require.NoError(t, err)
}

// TestCheckSpellingSuccess tests Check.Spelling success path
func TestCheckSpellingSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("misspell")

	err := Check{}.Spelling()
	require.NoError(t, err)
}

// TestCheckDocumentationSuccess tests Check.Documentation success path
func TestCheckDocumentationSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Check{}.Documentation()
	require.NoError(t, err)
}

// TestCheckDepsSuccess tests Check.Deps (alias) success path
func TestCheckDepsSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("go")

	err := Check{}.Deps()
	require.NoError(t, err)
}

// TestCheckDocsSuccess tests Check.Docs (alias) success path
func TestCheckDocsSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Check{}.Docs()
	require.NoError(t, err)
}

// TestCISetupSuccess tests CI.Setup success path with valid providers
func TestCISetupSuccess(t *testing.T) {
	providers := []string{"github", "gitlab", "jenkins", "circleci"}
	for _, provider := range providers {
		t.Run(provider, func(t *testing.T) {
			h := newOperationsTestHelper(t)
			defer h.teardown(t)
			h.expectCmd("echo")

			err := CI{}.Setup(provider)
			require.NoError(t, err)
		})
	}
}

// TestCISetupUnsupportedProvider tests CI.Setup with unsupported provider
func TestCISetupUnsupportedProvider(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)

	err := CI{}.Setup("unsupported")
	require.Error(t, err)
	assert.ErrorIs(t, err, errUnsupportedCIProvider)
}

// TestCIValidateSuccess tests CI.Validate success path
func TestCIValidateSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := CI{}.Validate()
	require.NoError(t, err)
}

// TestCIRunSuccess tests CI.Run success path
func TestCIRunSuccess(t *testing.T) {
	tests := []struct {
		name string
		job  string
	}{
		{"with job", "test"},
		{"empty job runs all", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newOperationsTestHelper(t)
			defer h.teardown(t)
			h.expectCmd("echo")

			err := CI{}.Run(tt.job)
			require.NoError(t, err)
		})
	}
}

// TestCIStatusSuccess tests CI.Status success path
func TestCIStatusSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := CI{}.Status("main")
	require.NoError(t, err)
}

// TestCILogsSuccess tests CI.Logs success path
func TestCILogsSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := CI{}.Logs("build-123")
	require.NoError(t, err)
}

// TestCITriggerSuccess tests CI.Trigger success path
func TestCITriggerSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := CI{}.Trigger("main", "ci")
	require.NoError(t, err)
}

// TestCISecretsSuccess tests CI.Secrets success path
func TestCISecretsSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := CI{}.Secrets("set", "API_KEY", "secret")
	require.NoError(t, err)
}

// TestCICacheSuccess tests CI.Cache success path
func TestCICacheSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := CI{}.Cache("clear")
	require.NoError(t, err)
}

// TestCIMatrixSuccess tests CI.Matrix success path
func TestCIMatrixSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	matrix := map[string][]string{"go": {"1.21", "1.22"}}
	err := CI{}.Matrix(matrix)
	require.NoError(t, err)
}

// TestCIArtifactsSuccess tests CI.Artifacts success path
func TestCIArtifactsSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := CI{}.Artifacts("download", "build-123")
	require.NoError(t, err)
}

// TestCIEnvironmentsSuccess tests CI.Environments success path
func TestCIEnvironmentsSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := CI{}.Environments("list", "production")
	require.NoError(t, err)
}

// TestMonitorStartSuccess tests Monitor.Start success path
func TestMonitorStartSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Monitor{}.Start()
	require.NoError(t, err)
}

// TestMonitorStopSuccess tests Monitor.Stop success path
func TestMonitorStopSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Monitor{}.Stop()
	require.NoError(t, err)
}

// TestMonitorStatusSuccess tests Monitor.Status success path
func TestMonitorStatusSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Monitor{}.Status()
	require.NoError(t, err)
}

// TestMonitorLogsSuccess tests Monitor.Logs success path
func TestMonitorLogsSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Monitor{}.Logs("api-service")
	require.NoError(t, err)
}

// TestMonitorMetricsSuccess tests Monitor.Metrics success path
func TestMonitorMetricsSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Monitor{}.Metrics("1h")
	require.NoError(t, err)
}

// TestMonitorAlertsSuccess tests Monitor.Alerts success path
func TestMonitorAlertsSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Monitor{}.Alerts()
	require.NoError(t, err)
}

// TestMonitorHealthSuccess tests Monitor.Health success path
func TestMonitorHealthSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Monitor{}.Health("http://localhost:8080/health")
	require.NoError(t, err)
}

// TestMonitorDashboardSuccess tests Monitor.Dashboard success path
func TestMonitorDashboardSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Monitor{}.Dashboard(8080)
	require.NoError(t, err)
}

// TestMonitorTraceSuccess tests Monitor.Trace success path
func TestMonitorTraceSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Monitor{}.Trace("trace-abc123")
	require.NoError(t, err)
}

// TestMonitorProfileSuccess tests Monitor.Profile success path
func TestMonitorProfileSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Monitor{}.Profile("cpu", "30s")
	require.NoError(t, err)
}

// TestMonitorExportSuccess tests Monitor.Export success path
func TestMonitorExportSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Monitor{}.Export("json", "24h")
	require.NoError(t, err)
}

// TestDatabaseMigrateSuccess tests Database.Migrate success path
func TestDatabaseMigrateSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Database{}.Migrate("up")
	require.NoError(t, err)
}

// TestDatabaseSeedSuccess tests Database.Seed success path
func TestDatabaseSeedSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Database{}.Seed("test_data")
	require.NoError(t, err)
}

// TestDatabaseResetSuccess tests Database.Reset success path
func TestDatabaseResetSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Database{}.Reset()
	require.NoError(t, err)
}

// TestDatabaseBackupSuccess tests Database.Backup success path
func TestDatabaseBackupSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Database{}.Backup("backup.sql")
	require.NoError(t, err)
}

// TestDatabaseRestoreSuccess tests Database.Restore success path
func TestDatabaseRestoreSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Database{}.Restore("backup.sql")
	require.NoError(t, err)
}

// TestDatabaseStatusSuccess tests Database.Status success path
func TestDatabaseStatusSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Database{}.Status()
	require.NoError(t, err)
}

// TestDatabaseCreateSuccess tests Database.Create success path
func TestDatabaseCreateSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Database{}.Create("mydb")
	require.NoError(t, err)
}

// TestDatabaseDropSuccess tests Database.Drop success path
func TestDatabaseDropSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Database{}.Drop("mydb")
	require.NoError(t, err)
}

// TestDatabaseConsoleSuccess tests Database.Console success path
func TestDatabaseConsoleSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Database{}.Console()
	require.NoError(t, err)
}

// TestDatabaseQuerySuccess tests Database.Query success path
func TestDatabaseQuerySuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Database{}.Query("SELECT * FROM users")
	require.NoError(t, err)
}

// TestDeployLocalSuccess tests Deploy.Local success path
func TestDeployLocalSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Deploy{}.Local()
	require.NoError(t, err)
}

// TestDeployStagingSuccess tests Deploy.Staging success path
func TestDeployStagingSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Deploy{}.Staging()
	require.NoError(t, err)
}

// TestDeployProductionSuccess tests Deploy.Production success path
func TestDeployProductionSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Deploy{}.Production()
	require.NoError(t, err)
}

// TestDeployKubernetesSuccess tests Deploy.Kubernetes success path
func TestDeployKubernetesSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Deploy{}.Kubernetes("production")
	require.NoError(t, err)
}

// TestDeployAWSSuccess tests Deploy.AWS success path
func TestDeployAWSSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Deploy{}.AWS("lambda")
	require.NoError(t, err)
}

// TestDeployGCPSuccess tests Deploy.GCP success path
func TestDeployGCPSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Deploy{}.GCP("cloud-run")
	require.NoError(t, err)
}

// TestDeployAzureSuccess tests Deploy.Azure success path
func TestDeployAzureSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Deploy{}.Azure("app-service")
	require.NoError(t, err)
}

// TestDeployHerokuSuccess tests Deploy.Heroku success path
func TestDeployHerokuSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Deploy{}.Heroku("myapp")
	require.NoError(t, err)
}

// TestDeployRollbackSuccess tests Deploy.Rollback success path
func TestDeployRollbackSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Deploy{}.Rollback("production", "v1.0.0")
	require.NoError(t, err)
}

// TestDeployStatusSuccess tests Deploy.Status success path
func TestDeployStatusSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Deploy{}.Status("production")
	require.NoError(t, err)
}

// TestCleanAllSuccess tests Clean.All success path
func TestCleanAllSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Clean{}.All()
	require.NoError(t, err)
}

// TestCleanBuildSuccess tests Clean.Build success path
func TestCleanBuildSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("go")

	err := Clean{}.Build()
	require.NoError(t, err)
}

// TestCleanBuildError tests Clean.Build error path
func TestCleanBuildError(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmdError("go")

	err := Clean{}.Build()
	require.Error(t, err)
}

// TestCleanTestSuccess tests Clean.Test success path
func TestCleanTestSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("go")

	err := Clean{}.Test()
	require.NoError(t, err)
}

// TestCleanCacheSuccess tests Clean.Cache success path
func TestCleanCacheSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("go")

	err := Clean{}.Cache()
	require.NoError(t, err)
}

// TestCleanDependenciesSuccess tests Clean.Dependencies success path
func TestCleanDependenciesSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("go")

	err := Clean{}.Dependencies()
	require.NoError(t, err)
}

// TestCleanDepsSuccess tests Clean.Deps success path
func TestCleanDepsSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("go")

	err := Clean{}.Deps()
	require.NoError(t, err)
}

// TestCleanFullSuccess tests Clean.Full success path
func TestCleanFullSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("go")

	err := Clean{}.Full()
	require.NoError(t, err)
}

// TestCleanFullFirstCmdError tests Clean.Full error on first command
func TestCleanFullFirstCmdError(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmdError("go")

	err := Clean{}.Full()
	require.Error(t, err)
}

// TestCleanGeneratedSuccess tests Clean.Generated success path
func TestCleanGeneratedSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("rm")

	err := Clean{}.Generated()
	require.NoError(t, err)
}

// TestCleanDistSuccess tests Clean.Dist success path
func TestCleanDistSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("rm")

	err := Clean{}.Dist()
	require.NoError(t, err)
}

// TestCleanLogsSuccess tests Clean.Logs success path
func TestCleanLogsSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("rm")

	err := Clean{}.Logs()
	require.NoError(t, err)
}

// TestCleanTempSuccess tests Clean.Temp success path
func TestCleanTempSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("rm")

	err := Clean{}.Temp()
	require.NoError(t, err)
}

// TestRunDevSuccess tests Run.Dev success path
func TestRunDevSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Run{}.Dev()
	require.NoError(t, err)
}

// TestRunProdSuccess tests Run.Prod success path
func TestRunProdSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Run{}.Prod()
	require.NoError(t, err)
}

// TestRunWatchSuccess tests Run.Watch success path
func TestRunWatchSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Run{}.Watch()
	require.NoError(t, err)
}

// TestRunDebugSuccess tests Run.Debug success path
func TestRunDebugSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Run{}.Debug()
	require.NoError(t, err)
}

// TestRunProfileSuccess tests Run.Profile success path
func TestRunProfileSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Run{}.Profile()
	require.NoError(t, err)
}

// TestRunBenchmarkSuccess tests Run.Benchmark success path
func TestRunBenchmarkSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Run{}.Benchmark()
	require.NoError(t, err)
}

// TestRunServerSuccess tests Run.Server success path
func TestRunServerSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Run{}.Server()
	require.NoError(t, err)
}

// TestRunMigrationsSuccess tests Run.Migrations success path
func TestRunMigrationsSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Run{}.Migrations()
	require.NoError(t, err)
}

// TestRunSeedsSuccess tests Run.Seeds success path
func TestRunSeedsSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Run{}.Seeds()
	require.NoError(t, err)
}

// TestRunWorkerSuccess tests Run.Worker success path
func TestRunWorkerSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Run{}.Worker()
	require.NoError(t, err)
}

// TestServeHTTPSuccess tests Serve.HTTP success path
func TestServeHTTPSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Serve{}.HTTP()
	require.NoError(t, err)
}

// TestServeHTTPSSuccess tests Serve.HTTPS success path
func TestServeHTTPSSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Serve{}.HTTPS()
	require.NoError(t, err)
}

// TestServeDocsSuccess tests Serve.Docs success path
func TestServeDocsSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Serve{}.Docs()
	require.NoError(t, err)
}

// TestServeAPISuccess tests Serve.API success path
func TestServeAPISuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Serve{}.API()
	require.NoError(t, err)
}

// TestServeGRPCSuccess tests Serve.GRPC success path
func TestServeGRPCSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Serve{}.GRPC()
	require.NoError(t, err)
}

// TestServeMetricsSuccess tests Serve.Metrics success path
func TestServeMetricsSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Serve{}.Metrics()
	require.NoError(t, err)
}

// TestServeStaticSuccess tests Serve.Static success path
func TestServeStaticSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Serve{}.Static()
	require.NoError(t, err)
}

// TestServeProxySuccess tests Serve.Proxy success path
func TestServeProxySuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Serve{}.Proxy()
	require.NoError(t, err)
}

// TestServeWebsocketSuccess tests Serve.Websocket success path
func TestServeWebsocketSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Serve{}.Websocket()
	require.NoError(t, err)
}

// TestServeWebSocketAliasSuccess tests Serve.WebSocket (alias) success path
func TestServeWebSocketAliasSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Serve{}.WebSocket()
	require.NoError(t, err)
}

// TestServeHealthCheckSuccess tests Serve.HealthCheck success path
func TestServeHealthCheckSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Serve{}.HealthCheck()
	require.NoError(t, err)
}

// TestCommonVersionSuccess tests Common.Version success path
func TestCommonVersionSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Common{}.Version()
	require.NoError(t, err)
}

// TestCommonDurationSuccess tests Common.Duration success path
func TestCommonDurationSuccess(t *testing.T) {
	h := newOperationsTestHelper(t)
	defer h.teardown(t)
	h.expectCmd("echo")

	err := Common{}.Duration()
	require.NoError(t, err)
}

// TestOperationsErrorPropagation tests error propagation for various operations
func TestOperationsErrorPropagation(t *testing.T) {
	tests := []struct {
		name string
		fn   func() error
		cmd  string
	}{
		{"Check.All", func() error { return Check{}.All() }, "echo"},
		{"CI.Validate", func() error { return CI{}.Validate() }, "echo"},
		{"Monitor.Start", func() error { return Monitor{}.Start() }, "echo"},
		{"Database.Reset", func() error { return Database{}.Reset() }, "echo"},
		{"Deploy.Local", func() error { return Deploy{}.Local() }, "echo"},
		{"Run.Dev", func() error { return Run{}.Dev() }, "echo"},
		{"Serve.HTTP", func() error { return Serve{}.HTTP() }, "echo"},
		{"Common.Version", func() error { return Common{}.Version() }, "echo"},
	}

	for _, tt := range tests {
		t.Run(tt.name+" error", func(t *testing.T) {
			h := newOperationsTestHelper(t)
			defer h.teardown(t)
			h.expectCmdError(tt.cmd)

			err := tt.fn()
			require.Error(t, err)
		})
	}
}
