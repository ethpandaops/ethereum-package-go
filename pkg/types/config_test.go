package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParticipantConfig(t *testing.T) {
	config := ParticipantConfig{
		ELType:         ClientGeth,
		CLType:         ClientLighthouse,
		ELVersion:      "v1.13.0",
		CLVersion:      "v4.5.0",
		Count:          2,
		ValidatorCount: 32,
	}

	assert.Equal(t, ClientGeth, config.ELType)
	assert.Equal(t, ClientLighthouse, config.CLType)
	assert.Equal(t, "v1.13.0", config.ELVersion)
	assert.Equal(t, "v4.5.0", config.CLVersion)
	assert.Equal(t, 2, config.Count)
	assert.Equal(t, 32, config.ValidatorCount)
}

func TestNetworkParams(t *testing.T) {
	params := NetworkParams{
		ChainID:                     12345,
		NetworkID:                   12345,
		SecondsPerSlot:              12,
		SlotsPerEpoch:               32,
		CapellaForkEpoch:            10,
		DenebForkEpoch:              20,
		ElectraForkEpoch:            30,
		MinValidatorWithdrawability: 256,
	}

	assert.Equal(t, uint64(12345), params.ChainID)
	assert.Equal(t, uint64(12345), params.NetworkID)
	assert.Equal(t, 12, params.SecondsPerSlot)
	assert.Equal(t, 32, params.SlotsPerEpoch)
	assert.Equal(t, 10, params.CapellaForkEpoch)
	assert.Equal(t, 20, params.DenebForkEpoch)
	assert.Equal(t, 30, params.ElectraForkEpoch)
	assert.Equal(t, 256, params.MinValidatorWithdrawability)
}

func TestMEVConfig(t *testing.T) {
	config := MEVConfig{
		Type:            "full",
		RelayURL:        "http://localhost:18550",
		MinBidEth:       "0.01",
		MaxBundleLength: 3,
	}

	assert.Equal(t, "full", config.Type)
	assert.Equal(t, "http://localhost:18550", config.RelayURL)
	assert.Equal(t, "0.01", config.MinBidEth)
	assert.Equal(t, 3, config.MaxBundleLength)
}

func TestAdditionalService(t *testing.T) {
	service := AdditionalService{
		Name: "prometheus",
		Config: map[string]interface{}{
			"port":     9090,
			"retention": "15d",
		},
	}

	assert.Equal(t, "prometheus", service.Name)
	assert.Equal(t, 9090, service.Config["port"])
	assert.Equal(t, "15d", service.Config["retention"])
}

func TestEthereumPackageConfig(t *testing.T) {
	config := EthereumPackageConfig{
		Participants: []ParticipantConfig{
			{
				ELType: ClientGeth,
				CLType: ClientLighthouse,
				Count:  2,
			},
		},
		NetworkParams: &NetworkParams{
			ChainID: 12345,
		},
		MEV: &MEVConfig{
			Type: "full",
		},
		AdditionalServices: []AdditionalService{
			{Name: "prometheus"},
		},
		GlobalClientLogLevel: "info",
	}

	assert.Len(t, config.Participants, 1)
	assert.NotNil(t, config.NetworkParams)
	assert.Equal(t, uint64(12345), config.NetworkParams.ChainID)
	assert.NotNil(t, config.MEV)
	assert.Equal(t, "full", config.MEV.Type)
	assert.Len(t, config.AdditionalServices, 1)
	assert.Equal(t, "info", config.GlobalClientLogLevel)
}

func TestPresetConfigSource(t *testing.T) {
	tests := []struct {
		name    string
		preset  Preset
		wantErr bool
	}{
		{"valid all-els", PresetAllELs, false},
		{"valid all-cls", PresetAllCLs, false},
		{"valid all-clients-matrix", PresetAllClientsMatrix, false},
		{"valid minimal", PresetMinimal, false},
		{"invalid preset", Preset("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := NewPresetConfigSource(tt.preset)
			assert.Equal(t, "preset", source.Type())

			err := source.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrInvalidPreset, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFileConfigSource(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid path", "/path/to/config.yaml", false},
		{"empty path", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := NewFileConfigSource(tt.path)
			assert.Equal(t, "file", source.Type())

			err := source.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrEmptyConfigPath, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInlineConfigSource(t *testing.T) {
	tests := []struct {
		name    string
		config  *EthereumPackageConfig
		wantErr bool
	}{
		{"valid config", &EthereumPackageConfig{}, false},
		{"nil config", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := NewInlineConfigSource(tt.config)
			assert.Equal(t, "inline", source.Type())

			err := source.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrNilConfig, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPresets(t *testing.T) {
	// Test that all preset constants are defined
	assert.Equal(t, Preset("all-els"), PresetAllELs)
	assert.Equal(t, Preset("all-cls"), PresetAllCLs)
	assert.Equal(t, Preset("all-clients-matrix"), PresetAllClientsMatrix)
	assert.Equal(t, Preset("minimal"), PresetMinimal)
}