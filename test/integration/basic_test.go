package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicSetup(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.Cleanup(t)

	assert.NotNil(t, env.Context)
	assert.NotNil(t, env.KurtosisClient)
}

func TestRunPackage(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.Cleanup(t)

	config := DefaultTestConfig()
	config.EnclaveName = GenerateTestEnclaveName(t)

	result, err := env.KurtosisClient.RunPackage(env.Context, config)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, config.EnclaveName, result.EnclaveName)
	assert.NotEmpty(t, result.ResponseLines)

	// Cleanup
	_ = env.KurtosisClient.DestroyEnclave(env.Context, config.EnclaveName)
}

func TestGetServices(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.Cleanup(t)

	config := DefaultTestConfig()
	config.EnclaveName = GenerateTestEnclaveName(t)

	// Run package first
	_, err := env.KurtosisClient.RunPackage(env.Context, config)
	require.NoError(t, err)

	// Get services
	services, err := env.KurtosisClient.GetServices(env.Context, config.EnclaveName)
	require.NoError(t, err)
	assert.NotEmpty(t, services)

	// Cleanup
	_ = env.KurtosisClient.DestroyEnclave(env.Context, config.EnclaveName)
}

func TestStopEnclave(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.Cleanup(t)

	config := DefaultTestConfig()
	config.EnclaveName = GenerateTestEnclaveName(t)

	// Run package first
	_, err := env.KurtosisClient.RunPackage(env.Context, config)
	require.NoError(t, err)

	// Stop enclave
	err = env.KurtosisClient.StopEnclave(env.Context, config.EnclaveName)
	assert.NoError(t, err)

	// Cleanup
	_ = env.KurtosisClient.DestroyEnclave(env.Context, config.EnclaveName)
}

func TestWaitForServices(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.Cleanup(t)

	config := DefaultTestConfig()
	config.EnclaveName = GenerateTestEnclaveName(t)

	// Run package first
	_, err := env.KurtosisClient.RunPackage(env.Context, config)
	require.NoError(t, err)

	// Wait for services (in mock mode, this succeeds immediately)
	err = env.KurtosisClient.WaitForServices(env.Context, config.EnclaveName, []string{"cl-1-geth-lighthouse"}, env.Timeout)
	assert.NoError(t, err)

	// Cleanup
	_ = env.KurtosisClient.DestroyEnclave(env.Context, config.EnclaveName)
}
