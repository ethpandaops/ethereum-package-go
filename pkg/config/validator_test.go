package config

import (
	"testing"

	"github.com/ethpandaops/ethereum-package-go/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestValidatorValidConfig(t *testing.T) {
	config := &types.EthereumPackageConfig{
		Participants: []types.ParticipantConfig{
			{
				ELType:         types.ClientGeth,
				CLType:         types.ClientLighthouse,
				Count:          2,
				ValidatorCount: 64,
			},
		},
		NetworkParams: &types.NetworkParams{
			ChainID:          12345,
			NetworkID:        12345,
			SecondsPerSlot:   12,
			SlotsPerEpoch:    32,
			CapellaForkEpoch: 10,
			DenebForkEpoch:   20,
			ElectraForkEpoch: 30,
		},
		MEV: &types.MEVConfig{
			Type:            "full",
			RelayURL:        "http://relay:18550",
			MaxBundleLength: 3,
		},
		AdditionalServices: []types.AdditionalService{
			{Name: "prometheus"},
			{Name: "grafana"},
		},
		GlobalClientLogLevel: "info",
	}

	validator := NewValidator(config)
	err := validator.Validate()
	assert.NoError(t, err)
}

func TestValidatorNilConfig(t *testing.T) {
	validator := NewValidator(nil)
	err := validator.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration is nil")
}

func TestValidatorParticipants(t *testing.T) {
	tests := []struct {
		name         string
		participants []types.ParticipantConfig
		wantErr      string
	}{
		{
			name:         "no participants",
			participants: []types.ParticipantConfig{},
			wantErr:      "at least one participant is required",
		},
		{
			name: "missing EL type",
			participants: []types.ParticipantConfig{
				{CLType: types.ClientLighthouse},
			},
			wantErr: "participant 0: execution layer type is required",
		},
		{
			name: "missing CL type",
			participants: []types.ParticipantConfig{
				{ELType: types.ClientGeth},
			},
			wantErr: "participant 0: consensus layer type is required",
		},
		{
			name: "invalid EL type",
			participants: []types.ParticipantConfig{
				{ELType: "invalid", CLType: types.ClientLighthouse},
			},
			wantErr: "participant 0: invalid execution layer type: invalid",
		},
		{
			name: "invalid CL type",
			participants: []types.ParticipantConfig{
				{ELType: types.ClientGeth, CLType: "invalid"},
			},
			wantErr: "participant 0: invalid consensus layer type: invalid",
		},
		{
			name: "negative count",
			participants: []types.ParticipantConfig{
				{ELType: types.ClientGeth, CLType: types.ClientLighthouse, Count: -1},
			},
			wantErr: "participant 0: count cannot be negative",
		},
		{
			name: "count too high",
			participants: []types.ParticipantConfig{
				{ELType: types.ClientGeth, CLType: types.ClientLighthouse, Count: 101},
			},
			wantErr: "participant 0: count cannot exceed 100",
		},
		{
			name: "negative validator count",
			participants: []types.ParticipantConfig{
				{ELType: types.ClientGeth, CLType: types.ClientLighthouse, ValidatorCount: -1},
			},
			wantErr: "participant 0: validator count cannot be negative",
		},
		{
			name: "validator count too high",
			participants: []types.ParticipantConfig{
				{ELType: types.ClientGeth, CLType: types.ClientLighthouse, ValidatorCount: 1000001},
			},
			wantErr: "participant 0: validator count cannot exceed 1000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &types.EthereumPackageConfig{
				Participants: tt.participants,
			}
			validator := NewValidator(config)
			err := validator.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestValidatorNetworkParams(t *testing.T) {
	tests := []struct {
		name    string
		params  *types.NetworkParams
		wantErr string
	}{
		{
			name: "invalid seconds per slot (too low)",
			params: &types.NetworkParams{
				SecondsPerSlot: -1,
			},
			wantErr: "seconds per slot must be between 1 and 120",
		},
		{
			name: "invalid seconds per slot (too high)",
			params: &types.NetworkParams{
				SecondsPerSlot: 121,
			},
			wantErr: "seconds per slot must be between 1 and 120",
		},
		{
			name: "invalid slots per epoch (too low)",
			params: &types.NetworkParams{
				SlotsPerEpoch: -1,
			},
			wantErr: "slots per epoch must be between 1 and 64",
		},
		{
			name: "invalid slots per epoch (too high)",
			params: &types.NetworkParams{
				SlotsPerEpoch: 65,
			},
			wantErr: "slots per epoch must be between 1 and 64",
		},
		{
			name: "negative fork epoch",
			params: &types.NetworkParams{
				CapellaForkEpoch: -1,
			},
			wantErr: "fork epochs cannot be negative",
		},
		{
			name: "invalid fork ordering (capella > deneb)",
			params: &types.NetworkParams{
				CapellaForkEpoch: 20,
				DenebForkEpoch:   10,
			},
			wantErr: "capella fork epoch must be before deneb fork epoch",
		},
		{
			name: "invalid fork ordering (deneb > electra)",
			params: &types.NetworkParams{
				DenebForkEpoch:   30,
				ElectraForkEpoch: 20,
			},
			wantErr: "deneb fork epoch must be before electra fork epoch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &types.EthereumPackageConfig{
				Participants: []types.ParticipantConfig{
					{ELType: types.ClientGeth, CLType: types.ClientLighthouse},
				},
				NetworkParams: tt.params,
			}
			validator := NewValidator(config)
			err := validator.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestValidatorMEV(t *testing.T) {
	tests := []struct {
		name    string
		mev     *types.MEVConfig
		wantErr string
	}{
		{
			name: "invalid MEV type",
			mev: &types.MEVConfig{
				Type: "invalid",
			},
			wantErr: "invalid MEV type: invalid",
		},
		{
			name: "invalid relay URL",
			mev: &types.MEVConfig{
				RelayURL: "not-a-url",
			},
			wantErr: "invalid MEV relay URL: not-a-url",
		},
		{
			name: "negative max bundle length",
			mev: &types.MEVConfig{
				MaxBundleLength: -1,
			},
			wantErr: "max bundle length cannot be negative",
		},
		{
			name: "max bundle length too high",
			mev: &types.MEVConfig{
				MaxBundleLength: 101,
			},
			wantErr: "max bundle length cannot exceed 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &types.EthereumPackageConfig{
				Participants: []types.ParticipantConfig{
					{ELType: types.ClientGeth, CLType: types.ClientLighthouse},
				},
				MEV: tt.mev,
			}
			validator := NewValidator(config)
			err := validator.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestValidatorAdditionalServices(t *testing.T) {
	tests := []struct {
		name     string
		services []types.AdditionalService
		wantErr  string
	}{
		{
			name: "missing service name",
			services: []types.AdditionalService{
				{Name: ""},
			},
			wantErr: "additional service 0: name is required",
		},
		{
			name: "duplicate service",
			services: []types.AdditionalService{
				{Name: "prometheus"},
				{Name: "prometheus"},
			},
			wantErr: "duplicate additional service: prometheus",
		},
		{
			name: "invalid service name",
			services: []types.AdditionalService{
				{Name: "invalid-service"},
			},
			wantErr: "invalid additional service name: invalid-service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &types.EthereumPackageConfig{
				Participants: []types.ParticipantConfig{
					{ELType: types.ClientGeth, CLType: types.ClientLighthouse},
				},
				AdditionalServices: tt.services,
			}
			validator := NewValidator(config)
			err := validator.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestValidatorGlobalSettings(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		wantErr  string
	}{
		{
			name:     "invalid log level",
			logLevel: "invalid",
			wantErr:  "invalid global client log level: invalid",
		},
		{
			name:     "valid log level lowercase",
			logLevel: "debug",
			wantErr:  "",
		},
		{
			name:     "valid log level uppercase",
			logLevel: "DEBUG",
			wantErr:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &types.EthereumPackageConfig{
				Participants: []types.ParticipantConfig{
					{ELType: types.ClientGeth, CLType: types.ClientLighthouse},
				},
				GlobalClientLogLevel: tt.logLevel,
			}
			validator := NewValidator(config)
			err := validator.Validate()
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatorHelperFunctions(t *testing.T) {
	// Test isValidELClient
	assert.True(t, isValidELClient(types.ClientGeth))
	assert.True(t, isValidELClient(types.ClientBesu))
	assert.True(t, isValidELClient(types.ClientNethermind))
	assert.True(t, isValidELClient(types.ClientErigon))
	assert.True(t, isValidELClient(types.ClientReth))
	assert.False(t, isValidELClient("invalid"))

	// Test isValidCLClient
	assert.True(t, isValidCLClient(types.ClientLighthouse))
	assert.True(t, isValidCLClient(types.ClientTeku))
	assert.True(t, isValidCLClient(types.ClientPrysm))
	assert.True(t, isValidCLClient(types.ClientNimbus))
	assert.True(t, isValidCLClient(types.ClientLodestar))
	assert.True(t, isValidCLClient(types.ClientGrandine))
	assert.False(t, isValidCLClient("invalid"))

	// Test isValidMEVType
	assert.True(t, isValidMEVType("none"))
	assert.True(t, isValidMEVType("mock"))
	assert.True(t, isValidMEVType("full"))
	assert.True(t, isValidMEVType("relay"))
	assert.False(t, isValidMEVType("invalid"))

	// Test isValidURL
	assert.True(t, isValidURL("http://example.com"))
	assert.True(t, isValidURL("https://example.com"))
	assert.False(t, isValidURL("ftp://example.com"))
	assert.False(t, isValidURL("example.com"))

	// Test isValidServiceName
	assert.True(t, isValidServiceName("prometheus"))
	assert.True(t, isValidServiceName("grafana"))
	assert.True(t, isValidServiceName("blockscout"))
	assert.False(t, isValidServiceName("invalid-service"))

	// Test isValidLogLevel
	assert.True(t, isValidLogLevel("trace"))
	assert.True(t, isValidLogLevel("debug"))
	assert.True(t, isValidLogLevel("info"))
	assert.True(t, isValidLogLevel("warn"))
	assert.True(t, isValidLogLevel("error"))
	assert.True(t, isValidLogLevel("fatal"))
	assert.True(t, isValidLogLevel("DEBUG"))
	assert.False(t, isValidLogLevel("invalid"))
}