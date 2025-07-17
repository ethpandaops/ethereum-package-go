package ethereum

import (
	"testing"
	"time"

	"github.com/ethpandaops/ethereum-package-go/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithPreset(t *testing.T) {
	tests := []struct {
		name   string
		preset config.Preset
	}{
		{"minimal preset", config.PresetMinimal},
		{"all ELs preset", config.PresetAllELs},
		{"all CLs preset", config.PresetAllCLs},
		{"all clients matrix", config.PresetAllClientsMatrix},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultRunConfig()
			opt := WithPreset(tt.preset)
			opt(cfg)

			require.NotNil(t, cfg.ConfigSource)
			assert.Equal(t, "preset", cfg.ConfigSource.Type())

			presetSource, ok := cfg.ConfigSource.(*config.PresetConfigSource)
			require.True(t, ok)
			assert.Equal(t, tt.preset, presetSource.GetPreset())
		})
	}
}

func TestWithConfigFile(t *testing.T) {
	cfg := defaultRunConfig()
	path := "/path/to/config.yaml"

	opt := WithConfigFile(path)
	opt(cfg)

	require.NotNil(t, cfg.ConfigSource)
	assert.Equal(t, "file", cfg.ConfigSource.Type())

	fileSource, ok := cfg.ConfigSource.(*config.FileConfigSource)
	require.True(t, ok)
	assert.Equal(t, path, fileSource.GetPath())
}

func TestWithConfig(t *testing.T) {
	cfg := defaultRunConfig()
	ethConfig := &config.EthereumPackageConfig{
		Participants: []config.ParticipantConfig{
			{ELType: "geth", CLType: "lighthouse", Count: 1},
		},
	}

	opt := WithConfig(ethConfig)
	opt(cfg)

	require.NotNil(t, cfg.ConfigSource)
	assert.Equal(t, "inline", cfg.ConfigSource.Type())

	inlineSource, ok := cfg.ConfigSource.(*config.InlineConfigSource)
	require.True(t, ok)
	assert.Equal(t, ethConfig, inlineSource.GetConfig())
}

func TestWithChainID(t *testing.T) {
	cfg := defaultRunConfig()
	chainID := uint64(99999)

	opt := WithChainID(chainID)
	opt(cfg)

	assert.Equal(t, chainID, cfg.ChainID)
}

func TestWithNetworkParams(t *testing.T) {
	cfg := defaultRunConfig()
	params := &config.NetworkParams{
		Network:                 "kurtosis",
		NetworkID:               "12345",
		SecondsPerSlot:          12,
		NumValidatorKeysPerNode: 64,
	}

	opt := WithNetworkParams(params)
	opt(cfg)

	assert.Equal(t, params, cfg.NetworkParams)
}

func TestWithMEV(t *testing.T) {
	cfg := defaultRunConfig()
	mevConfig := &config.MEVConfig{
		Type:     "full",
		RelayURL: "http://relay.example.com",
	}

	opt := WithMEV(mevConfig)
	opt(cfg)

	assert.Equal(t, mevConfig, cfg.MEV)
}

func TestWithAdditionalServices(t *testing.T) {
	cfg := defaultRunConfig()
	services := []string{"prometheus", "grafana", "dora"}

	opt := WithAdditionalServices(services...)
	opt(cfg)

	require.Len(t, cfg.AdditionalServices, 3)
	for i, service := range services {
		assert.Equal(t, service, string(cfg.AdditionalServices[i]))
	}
}

func TestWithAdditionalService(t *testing.T) {
	cfg := defaultRunConfig()
	service := config.AdditionalService("prometheus")

	opt := WithAdditionalService(service)
	opt(cfg)

	require.Len(t, cfg.AdditionalServices, 1)
	assert.Equal(t, service, cfg.AdditionalServices[0])
}

func TestWithGlobalLogLevel(t *testing.T) {
	cfg := defaultRunConfig()
	logLevel := "debug"

	opt := WithGlobalLogLevel(logLevel)
	opt(cfg)

	assert.Equal(t, logLevel, cfg.GlobalLogLevel)
}

func TestWithEnclaveName(t *testing.T) {
	cfg := defaultRunConfig()
	name := "test-enclave"

	opt := WithEnclaveName(name)
	opt(cfg)

	assert.Equal(t, name, cfg.EnclaveName)
}

func TestWithPackageID(t *testing.T) {
	cfg := defaultRunConfig()
	packageID := "github.com/custom/package"

	opt := WithPackageID(packageID)
	opt(cfg)

	assert.Equal(t, packageID, cfg.PackageID)
}

func TestWithPackageVersion(t *testing.T) {
	cfg := defaultRunConfig()
	version := "2.5.0"

	opt := WithPackageVersion(version)
	opt(cfg)

	assert.Equal(t, version, cfg.PackageVersion)
}

func TestWithPackageRepo(t *testing.T) {
	cfg := defaultRunConfig()
	repo := "github.com/custom/ethereum-package"
	version := "1.2.3"

	opt := WithPackageRepo(repo, version)
	opt(cfg)

	assert.Equal(t, repo, cfg.PackageID)
	assert.Equal(t, version, cfg.PackageVersion)
}

func TestWithDryRun(t *testing.T) {
	cfg := defaultRunConfig()

	opt := WithDryRun(true)
	opt(cfg)

	assert.True(t, cfg.DryRun)
}

func TestWithParallelism(t *testing.T) {
	cfg := defaultRunConfig()
	parallelism := 8

	opt := WithParallelism(parallelism)
	opt(cfg)

	assert.Equal(t, parallelism, cfg.Parallelism)
}

func TestWithVerbose(t *testing.T) {
	cfg := defaultRunConfig()

	opt := WithVerbose(true)
	opt(cfg)

	assert.True(t, cfg.VerboseMode)
}

func TestWithTimeout(t *testing.T) {
	cfg := defaultRunConfig()
	timeout := 30 * time.Minute

	opt := WithTimeout(timeout)
	opt(cfg)

	assert.Equal(t, timeout, cfg.Timeout)
}

func TestConvenienceOptions(t *testing.T) {
	tests := []struct {
		name     string
		optFunc  RunOption
		validate func(t *testing.T, cfg *RunConfig)
	}{
		{
			name:    "AllELs",
			optFunc: AllELs(),
			validate: func(t *testing.T, cfg *RunConfig) {
				presetSource, ok := cfg.ConfigSource.(*config.PresetConfigSource)
				require.True(t, ok)
				assert.Equal(t, config.PresetAllELs, presetSource.GetPreset())
			},
		},
		{
			name:    "AllCLs",
			optFunc: AllCLs(),
			validate: func(t *testing.T, cfg *RunConfig) {
				presetSource, ok := cfg.ConfigSource.(*config.PresetConfigSource)
				require.True(t, ok)
				assert.Equal(t, config.PresetAllCLs, presetSource.GetPreset())
			},
		},
		{
			name:    "AllClientsMatrix",
			optFunc: AllClientsMatrix(),
			validate: func(t *testing.T, cfg *RunConfig) {
				presetSource, ok := cfg.ConfigSource.(*config.PresetConfigSource)
				require.True(t, ok)
				assert.Equal(t, config.PresetAllClientsMatrix, presetSource.GetPreset())
			},
		},
		{
			name:    "Minimal",
			optFunc: Minimal(),
			validate: func(t *testing.T, cfg *RunConfig) {
				presetSource, ok := cfg.ConfigSource.(*config.PresetConfigSource)
				require.True(t, ok)
				assert.Equal(t, config.PresetMinimal, presetSource.GetPreset())
			},
		},
		{
			name:    "WithExplorer",
			optFunc: WithExplorer(),
			validate: func(t *testing.T, cfg *RunConfig) {
				require.Len(t, cfg.AdditionalServices, 1)
				assert.Equal(t, "dora", string(cfg.AdditionalServices[0]))
			},
		},
		{
			name:    "WithFullObservability",
			optFunc: WithFullObservability(),
			validate: func(t *testing.T, cfg *RunConfig) {
				require.Len(t, cfg.AdditionalServices, 3)
				assert.Equal(t, "prometheus", string(cfg.AdditionalServices[0]))
				assert.Equal(t, "grafana", string(cfg.AdditionalServices[1]))
				assert.Equal(t, "dora", string(cfg.AdditionalServices[2]))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultRunConfig()
			tt.optFunc(cfg)
			tt.validate(t, cfg)
		})
	}
}

func TestWithMEVBoost(t *testing.T) {
	cfg := defaultRunConfig()

	opt := WithMEVBoost()
	opt(cfg)

	require.NotNil(t, cfg.MEV)
	assert.Equal(t, "full", cfg.MEV.Type)
}

func TestWithMEVBoostRelay(t *testing.T) {
	cfg := defaultRunConfig()
	relayURL := "http://custom-relay.example.com"

	opt := WithMEVBoostRelay(relayURL)
	opt(cfg)

	require.NotNil(t, cfg.MEV)
	assert.Equal(t, "full", cfg.MEV.Type)
	assert.Equal(t, relayURL, cfg.MEV.RelayURL)
}

func TestMultipleOptions(t *testing.T) {
	cfg := defaultRunConfig()

	// Apply multiple options
	opts := []RunOption{
		WithChainID(99999),
		WithAdditionalServices("prometheus", "grafana"),
		WithVerbose(true),
		WithTimeout(20 * time.Minute),
		WithGlobalLogLevel("debug"),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	assert.Equal(t, uint64(99999), cfg.ChainID)
	assert.Len(t, cfg.AdditionalServices, 2)
	assert.True(t, cfg.VerboseMode)
	assert.Equal(t, 20*time.Minute, cfg.Timeout)
	assert.Equal(t, "debug", cfg.GlobalLogLevel)
}

func TestWithPortPublisher(t *testing.T) {
	cfg := defaultRunConfig()
	portPublisher := &config.PortPublisherConfig{
		NatExitIP: "192.168.1.100",
		EL: &config.PortPublisherComponent{
			Enabled:         true,
			PublicPortStart: 32000,
		},
		CL: &config.PortPublisherComponent{
			Enabled:         true,
			PublicPortStart: 33000,
		},
		VC: &config.PortPublisherComponent{
			Enabled:         true,
			PublicPortStart: 34000,
		},
	}

	opt := WithPortPublisher(portPublisher)
	opt(cfg)

	assert.Equal(t, portPublisher, cfg.PortPublisher)
	assert.Equal(t, "192.168.1.100", cfg.PortPublisher.NatExitIP)
	assert.True(t, cfg.PortPublisher.EL.Enabled)
	assert.Equal(t, 32000, cfg.PortPublisher.EL.PublicPortStart)
}

func TestWithNATExitIP(t *testing.T) {
	tests := []struct {
		name              string
		ip                string
		existingConfig    *config.PortPublisherConfig
		expectedNatExitIP string
		expectedELEnabled bool
		expectedCLEnabled bool
	}{
		{
			name:              "set NAT exit IP on nil config",
			ip:                "127.0.0.1",
			existingConfig:    nil,
			expectedNatExitIP: "127.0.0.1",
			expectedELEnabled: true,
			expectedCLEnabled: true,
		},
		{
			name:              "set NAT exit IP on empty config",
			ip:                "192.168.1.100",
			existingConfig:    &config.PortPublisherConfig{},
			expectedNatExitIP: "192.168.1.100",
			expectedELEnabled: true,
			expectedCLEnabled: true,
		},
		{
			name:              "override existing NAT exit IP",
			ip:                "10.0.0.1",
			existingConfig:    &config.PortPublisherConfig{NatExitIP: "192.168.1.1"},
			expectedNatExitIP: "10.0.0.1",
			expectedELEnabled: true,
			expectedCLEnabled: true,
		},
		{
			name: "preserve existing component settings",
			ip:   "172.16.0.1",
			existingConfig: &config.PortPublisherConfig{
				EL: &config.PortPublisherComponent{Enabled: false},
				CL: &config.PortPublisherComponent{Enabled: false},
			},
			expectedNatExitIP: "172.16.0.1",
			expectedELEnabled: false,
			expectedCLEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultRunConfig()
			cfg.PortPublisher = tt.existingConfig

			opt := WithNATExitIP(tt.ip)
			opt(cfg)

			require.NotNil(t, cfg.PortPublisher)
			assert.Equal(t, tt.expectedNatExitIP, cfg.PortPublisher.NatExitIP)

			if cfg.PortPublisher.EL != nil {
				assert.Equal(t, tt.expectedELEnabled, cfg.PortPublisher.EL.Enabled)
			}
			if cfg.PortPublisher.CL != nil {
				assert.Equal(t, tt.expectedCLEnabled, cfg.PortPublisher.CL.Enabled)
			}
		})
	}
}

func TestPortPublisherWithOtherOptions(t *testing.T) {
	cfg := defaultRunConfig()

	// Apply multiple options including port publisher
	opts := []RunOption{
		WithChainID(12345),
		WithNATExitIP("127.0.0.1"),
		WithAdditionalServices("prometheus"),
		WithGlobalLogLevel("debug"),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	// Verify all options were applied
	assert.Equal(t, uint64(12345), cfg.ChainID)
	assert.Equal(t, "debug", cfg.GlobalLogLevel)
	assert.Len(t, cfg.AdditionalServices, 1)

	// Verify port publisher was configured
	require.NotNil(t, cfg.PortPublisher)
	assert.Equal(t, "127.0.0.1", cfg.PortPublisher.NatExitIP)
	assert.True(t, cfg.PortPublisher.EL.Enabled)
	assert.True(t, cfg.PortPublisher.CL.Enabled)
}
