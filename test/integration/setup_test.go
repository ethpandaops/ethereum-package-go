package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ethpandaops/ethereum-package-go/pkg/kurtosis"
	"github.com/ethpandaops/ethereum-package-go/test/mocks"
)

// TestEnvironment holds the test environment configuration
type TestEnvironment struct {
	Context        context.Context
	Cancel         context.CancelFunc
	KurtosisClient kurtosis.Client
	UseMock        bool
	Timeout        time.Duration
}

// SetupTestEnvironment creates a new test environment
func SetupTestEnvironment(t *testing.T) *TestEnvironment {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	env := &TestEnvironment{
		Context: ctx,
		Cancel:  cancel,
		Timeout: 5 * time.Minute,
	}

	// Check if we should use real Kurtosis or mock
	if os.Getenv("INTEGRATION_TEST_REAL") == "true" {
		// Use real Kurtosis client
		client, err := kurtosis.NewKurtosisClient(ctx)
		if err != nil {
			t.Skipf("Skipping integration test: Kurtosis not available: %v", err)
		}
		env.KurtosisClient = client
		env.UseMock = false
	} else {
		// Use mock client
		env.KurtosisClient = mocks.NewMockKurtosisClient()
		env.UseMock = true
	}

	return env
}

// Cleanup cleans up the test environment
func (e *TestEnvironment) Cleanup(t *testing.T) {
	t.Helper()

	if e.Cancel != nil {
		e.Cancel()
	}
}

// SkipIfNotReal skips the test if not running against real Kurtosis
func (e *TestEnvironment) SkipIfNotReal(t *testing.T) {
	t.Helper()

	if e.UseMock {
		t.Skip("Skipping test that requires real Kurtosis")
	}
}

// SkipIfSlow skips the test if running in short mode
func SkipIfSlow(t *testing.T) {
	t.Helper()

	if testing.Short() {
		t.Skip("Skipping slow test in short mode")
	}
}

// RequireEnvVar checks that an environment variable is set
func RequireEnvVar(t *testing.T, name string) string {
	t.Helper()

	value := os.Getenv(name)
	if value == "" {
		t.Skipf("Skipping test: %s environment variable not set", name)
	}
	return value
}

// WaitForCondition waits for a condition to be true
func WaitForCondition(t *testing.T, timeout time.Duration, check func() bool, message string) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		if check() {
			return
		}
		<-ticker.C
	}

	t.Fatalf("Condition not met within timeout: %s", message)
}

// GenerateTestEnclaveName generates a unique enclave name for testing
func GenerateTestEnclaveName(t *testing.T) string {
	t.Helper()

	return "test-" + t.Name() + "-" + time.Now().Format("20060102-150405")
}

// DefaultTestConfig returns a default test configuration
func DefaultTestConfig() kurtosis.RunPackageConfig {
	return kurtosis.RunPackageConfig{
		PackageID:    "github.com/ethpandaops/ethereum-package",
		EnclaveName:  "test-enclave",
		ConfigYAML:   "participants:\n  - el_type: geth\n    cl_type: lighthouse\n    count: 1",
		DryRun:       false,
		Parallelism:  4,
		VerboseMode:  true,
	}
}