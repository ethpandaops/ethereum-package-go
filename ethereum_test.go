package ethereum

import (
	"context"
	"testing"
	"time"

	"github.com/ethpandaops/ethereum-package-go/pkg/client"
	"github.com/ethpandaops/ethereum-package-go/pkg/config"
	"github.com/ethpandaops/ethereum-package-go/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultRunConfig(t *testing.T) {
	cfg := defaultRunConfig()

	assert.Equal(t, DefaultPackageRepository, cfg.PackageID)
	assert.Equal(t, DefaultPackageVersion, cfg.PackageVersion)
	assert.NotEmpty(t, cfg.EnclaveName)
	assert.NotNil(t, cfg.ConfigSource)
	assert.Equal(t, uint64(12345), cfg.ChainID)
	assert.False(t, cfg.DryRun)
	assert.Equal(t, 4, cfg.Parallelism)
	assert.False(t, cfg.VerboseMode)
	assert.Equal(t, 10*time.Minute, cfg.Timeout)
	assert.Equal(t, "info", cfg.GlobalLogLevel)
}

func TestValidateRunConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *RunConfig
		wantErr string
	}{
		{
			name: "valid config",
			cfg: &RunConfig{
				PackageID:    "github.com/ethpandaops/ethereum-package",
				EnclaveName:  "test-enclave",
				ConfigSource: config.NewPresetConfigSource(config.PresetMinimal),
				Timeout:      time.Minute,
			},
		},
		{
			name: "missing package ID",
			cfg: &RunConfig{
				EnclaveName:  "test-enclave",
				ConfigSource: config.NewPresetConfigSource(config.PresetMinimal),
				Timeout:      time.Minute,
			},
			wantErr: "package ID is required",
		},
		{
			name: "missing enclave name",
			cfg: &RunConfig{
				PackageID:    "github.com/ethpandaops/ethereum-package",
				ConfigSource: config.NewPresetConfigSource(config.PresetMinimal),
				Timeout:      time.Minute,
			},
			wantErr: "enclave name is required",
		},
		{
			name: "missing config source",
			cfg: &RunConfig{
				PackageID:   "github.com/ethpandaops/ethereum-package",
				EnclaveName: "test-enclave",
				Timeout:     time.Minute,
			},
			wantErr: "config source is required",
		},
		{
			name: "invalid timeout",
			cfg: &RunConfig{
				PackageID:    "github.com/ethpandaops/ethereum-package",
				EnclaveName:  "test-enclave",
				ConfigSource: config.NewPresetConfigSource(config.PresetMinimal),
				Timeout:      0,
			},
			wantErr: "timeout must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRunConfig(tt.cfg)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildEthereumConfig(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *RunConfig
		validate func(*testing.T, *config.EthereumPackageConfig)
	}{
		{
			name: "preset config",
			cfg: &RunConfig{
				ConfigSource: config.NewPresetConfigSource(config.PresetMinimal),
				ChainID:      98765,
			},
			validate: func(t *testing.T, config *config.EthereumPackageConfig) {
				assert.Len(t, config.Participants, 1)
				assert.Equal(t, client.Geth, config.Participants[0].ELType)
				assert.Equal(t, client.Lighthouse, config.Participants[0].CLType)
				require.NotNil(t, config.NetworkParams)
				assert.Equal(t, "98765", config.NetworkParams.NetworkID)
			},
		},
		{
			name: "inline config",
			cfg: &RunConfig{
				ConfigSource: config.NewInlineConfigSource(&config.EthereumPackageConfig{
					Participants: []config.ParticipantConfig{
						{ELType: client.Besu, CLType: client.Teku, Count: 2},
					},
				}),
				MEV: &config.MEVConfig{Type: "full"},
			},
			validate: func(t *testing.T, config *config.EthereumPackageConfig) {
				assert.Len(t, config.Participants, 1)
				assert.Equal(t, client.Besu, config.Participants[0].ELType)
				assert.Equal(t, client.Teku, config.Participants[0].CLType)
				require.NotNil(t, config.MEV)
				assert.Equal(t, "full", config.MEV.Type)
			},
		},
		{
			name: "with additional services",
			cfg: &RunConfig{
				ConfigSource: config.NewPresetConfigSource(config.PresetMinimal),
				AdditionalServices: []config.AdditionalService{
					"prometheus",
					"grafana",
				},
				GlobalLogLevel: "debug",
			},
			validate: func(t *testing.T, config *config.EthereumPackageConfig) {
				assert.Len(t, config.AdditionalServices, 2)
				assert.Equal(t, "prometheus", string(config.AdditionalServices[0]))
				assert.Equal(t, "grafana", string(config.AdditionalServices[1]))
				assert.Equal(t, "debug", config.GlobalLogLevel)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := buildEthereumConfig(tt.cfg)
			require.NoError(t, err)
			require.NotNil(t, config)
			tt.validate(t, config)
		})
	}
}

func TestRunWithMockClient(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewMockKurtosisClient()

	network, err := Run(ctx,
		WithPreset(config.PresetMinimal),
		WithChainID(54321),
		WithEnclaveName("test-run-enclave"),
		WithKurtosisClient(mockClient),
		WithDryRun(true), // Use dry run to avoid service discovery
	)

	require.NoError(t, err)
	assert.NotNil(t, network)

	// Verify mock was called
	assert.Equal(t, 1, mockClient.CallCount["RunPackage"])
	assert.NotNil(t, mockClient.LastRunConfig)
	assert.Equal(t, "test-run-enclave", mockClient.LastRunConfig.EnclaveName)
	assert.True(t, mockClient.LastRunConfig.DryRun)
}

func TestRunConfigOptions(t *testing.T) {
	cfg := defaultRunConfig()

	// Test individual options
	WithPreset(config.PresetAllELs)(cfg)
	assert.Equal(t, "preset", cfg.ConfigSource.Type())

	WithChainID(99999)(cfg)
	assert.Equal(t, uint64(99999), cfg.ChainID)

	WithMEV(&config.MEVConfig{Type: "mock"})(cfg)
	assert.NotNil(t, cfg.MEV)
	assert.Equal(t, "mock", cfg.MEV.Type)

	WithAdditionalServices("prometheus", "grafana")(cfg)
	assert.Len(t, cfg.AdditionalServices, 2)

	WithGlobalLogLevel("trace")(cfg)
	assert.Equal(t, "trace", cfg.GlobalLogLevel)

	WithEnclaveName("custom-enclave")(cfg)
	assert.Equal(t, "custom-enclave", cfg.EnclaveName)

	WithDryRun(true)(cfg)
	assert.True(t, cfg.DryRun)

	WithParallelism(8)(cfg)
	assert.Equal(t, 8, cfg.Parallelism)

	WithVerbose(true)(cfg)
	assert.True(t, cfg.VerboseMode)

	WithTimeout(5 * time.Minute)(cfg)
	assert.Equal(t, 5*time.Minute, cfg.Timeout)
}

func TestConvenienceFunctions(t *testing.T) {
	cfg := defaultRunConfig()

	// Test convenience preset functions
	AllELs()(cfg)
	source := cfg.ConfigSource.(*config.PresetConfigSource)
	assert.Equal(t, config.PresetAllELs, source.GetPreset())

	AllCLs()(cfg)
	source = cfg.ConfigSource.(*config.PresetConfigSource)
	assert.Equal(t, config.PresetAllCLs, source.GetPreset())

	AllClientsMatrix()(cfg)
	source = cfg.ConfigSource.(*config.PresetConfigSource)
	assert.Equal(t, config.PresetAllClientsMatrix, source.GetPreset())

	Minimal()(cfg)
	source = cfg.ConfigSource.(*config.PresetConfigSource)
	assert.Equal(t, config.PresetMinimal, source.GetPreset())

	// Test explorer function
	cfg.AdditionalServices = nil
	WithExplorer()(cfg)
	assert.Len(t, cfg.AdditionalServices, 1)
	assert.Equal(t, "dora", string(cfg.AdditionalServices[0]))

	cfg.AdditionalServices = nil
	WithFullObservability()(cfg)
	assert.Len(t, cfg.AdditionalServices, 3)

	// Test custom chain
	WithCustomChain("777", 6, 64)(cfg)
	require.NotNil(t, cfg.NetworkParams)
	assert.Equal(t, "777", cfg.NetworkParams.NetworkID)
	assert.Equal(t, 6, cfg.NetworkParams.SecondsPerSlot)
	assert.Equal(t, 64, cfg.NetworkParams.NumValidatorKeysPerNode)

	// Test MEV functions
	WithMEVBoost()(cfg)
	require.NotNil(t, cfg.MEV)
	assert.Equal(t, "full", cfg.MEV.Type)

	WithMEVBoostRelay("http://relay:18550")(cfg)
	require.NotNil(t, cfg.MEV)
	assert.Equal(t, "full", cfg.MEV.Type)
	assert.Equal(t, "http://relay:18550", cfg.MEV.RelayURL)
}

func TestPackageVersionOptions(t *testing.T) {
	cfg := defaultRunConfig()

	// Test WithPackageVersion
	WithPackageVersion("5.0.0")(cfg)
	assert.Equal(t, "5.0.0", cfg.PackageVersion)

	// Test WithPackageRepo
	WithPackageRepo("github.com/custom/package", "1.0.0")(cfg)
	assert.Equal(t, "github.com/custom/package", cfg.PackageID)
	assert.Equal(t, "1.0.0", cfg.PackageVersion)
}

func TestPackageIDConstruction(t *testing.T) {
	tests := []struct {
		name           string
		packageID      string
		packageVersion string
		expectedID     string
	}{
		{
			name:           "with version",
			packageID:      "github.com/ethpandaops/ethereum-package",
			packageVersion: "5.0.1",
			expectedID:     "github.com/ethpandaops/ethereum-package@5.0.1",
		},
		{
			name:           "without version",
			packageID:      "github.com/ethpandaops/ethereum-package",
			packageVersion: "",
			expectedID:     "github.com/ethpandaops/ethereum-package",
		},
		{
			name:           "custom repo with version",
			packageID:      "github.com/custom/ethereum-package",
			packageVersion: "2.0.0",
			expectedID:     "github.com/custom/ethereum-package@2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the package ID construction logic
			packageID := tt.packageID
			if tt.packageVersion != "" {
				packageID = packageID + "@" + tt.packageVersion
			}
			assert.Equal(t, tt.expectedID, packageID)
		})
	}
}
