package config

import (
	"testing"

	"github.com/ethpandaops/ethereum-package-go/pkg/client"
	"github.com/stretchr/testify/assert"
)

// ValidatorTestCase represents a test case for config validation
type ValidatorTestCase struct {
	Name    string
	Config  *EthereumPackageConfig
	WantErr string
}

// RunValidatorTests runs a set of validator test cases
func RunValidatorTests(t *testing.T, tests []ValidatorTestCase) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			validator := NewValidator(tt.Config)
			err := validator.Validate()
			if tt.WantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.WantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// DefaultValidConfig returns a valid default configuration for testing
func DefaultValidConfig() *EthereumPackageConfig {
	return &EthereumPackageConfig{
		Participants: []ParticipantConfig{
			{
				ELType: client.Geth,
				CLType: client.Lighthouse,
				Count:  1,
			},
		},
	}
}

// DefaultValidConfigWithNetworkParams returns a valid config with network parameters
func DefaultValidConfigWithNetworkParams() *EthereumPackageConfig {
	config := DefaultValidConfig()
	config.NetworkParams = &NetworkParams{
		Network:                 "kurtosis",
		NetworkID:               "12345",
		SecondsPerSlot:          12,
		NumValidatorKeysPerNode: 64,
		GenesisDelay:            20,
		AltairForkEpoch:         0,
		BellatrixForkEpoch:      0,
		CapellaForkEpoch:        10,
		DenebForkEpoch:          20,
		ElectraForkEpoch:        30,
	}
	return config
}

// DefaultValidConfigWithMEV returns a valid config with MEV settings
func DefaultValidConfigWithMEV() *EthereumPackageConfig {
	config := DefaultValidConfig()
	config.MEV = &MEVConfig{
		Type:            "full",
		RelayURL:        "http://relay:18550",
		MaxBundleLength: 3,
	}
	return config
}

// ParticipantTestCases returns common test cases for participant validation
func ParticipantTestCases() []ValidatorTestCase {
	return []ValidatorTestCase{
		{
			Name:    "no participants",
			Config:  &EthereumPackageConfig{Participants: []ParticipantConfig{}},
			WantErr: "at least one participant is required",
		},
		{
			Name: "missing EL type",
			Config: &EthereumPackageConfig{
				Participants: []ParticipantConfig{{CLType: client.Lighthouse}},
			},
			WantErr: "participant 0: execution layer type is required",
		},
		{
			Name: "missing CL type",
			Config: &EthereumPackageConfig{
				Participants: []ParticipantConfig{{ELType: client.Geth}},
			},
			WantErr: "participant 0: consensus layer type is required",
		},
		{
			Name: "invalid EL type",
			Config: &EthereumPackageConfig{
				Participants: []ParticipantConfig{
					{ELType: "invalid", CLType: client.Lighthouse},
				},
			},
			WantErr: "participant 0: invalid execution client type: invalid",
		},
		{
			Name: "invalid CL type",
			Config: &EthereumPackageConfig{
				Participants: []ParticipantConfig{
					{ELType: client.Geth, CLType: "invalid"},
				},
			},
			WantErr: "participant 0: invalid consensus client type: invalid",
		},
		{
			Name: "negative count",
			Config: &EthereumPackageConfig{
				Participants: []ParticipantConfig{
					{ELType: client.Geth, CLType: client.Lighthouse, Count: -1},
				},
			},
			WantErr: "participant 0: count cannot be negative",
		},
		{
			Name: "count too high",
			Config: &EthereumPackageConfig{
				Participants: []ParticipantConfig{
					{ELType: client.Geth, CLType: client.Lighthouse, Count: 101},
				},
			},
			WantErr: "participant 0: count 101 exceeds maximum of 100",
		},
	}
}

// NetworkParamsTestCases returns common test cases for network params validation
func NetworkParamsTestCases() []ValidatorTestCase {
	baseConfig := DefaultValidConfig()

	createConfigWithNetworkParams := func(params *NetworkParams) *EthereumPackageConfig {
		config := &EthereumPackageConfig{
			Participants:  baseConfig.Participants,
			NetworkParams: params,
		}
		return config
	}

	return []ValidatorTestCase{
		{
			Name: "invalid seconds per slot (too low)",
			Config: createConfigWithNetworkParams(&NetworkParams{
				SecondsPerSlot: -1,
			}),
			WantErr: "seconds per slot must be between 1 and 60",
		},
		{
			Name: "invalid seconds per slot (too high)",
			Config: createConfigWithNetworkParams(&NetworkParams{
				SecondsPerSlot: 121,
			}),
			WantErr: "seconds per slot must be between 1 and 60",
		},
		{
			Name: "invalid validator keys per node (too low)",
			Config: createConfigWithNetworkParams(&NetworkParams{
				SecondsPerSlot:          12,
				NumValidatorKeysPerNode: -1,
			}),
			WantErr: "num validator keys per node must be between 0 and 1000000",
		},
		{
			Name: "invalid validator keys per node (too high)",
			Config: createConfigWithNetworkParams(&NetworkParams{
				SecondsPerSlot:          12,
				NumValidatorKeysPerNode: 1000001,
			}),
			WantErr: "num validator keys per node must be between 0 and 1000000",
		},
		{
			Name: "negative genesis delay",
			Config: createConfigWithNetworkParams(&NetworkParams{
				SecondsPerSlot: 12,
				GenesisDelay:   -1,
			}),
			WantErr: "genesis delay cannot be negative",
		},
		{
			Name: "negative fork epoch",
			Config: createConfigWithNetworkParams(&NetworkParams{
				SecondsPerSlot:   12,
				CapellaForkEpoch: -1,
			}),
			WantErr: "fork epochs cannot be negative",
		},
		{
			Name: "invalid fork ordering",
			Config: createConfigWithNetworkParams(&NetworkParams{
				SecondsPerSlot:   12,
				CapellaForkEpoch: 20,
				DenebForkEpoch:   10,
			}),
			WantErr: "fork epochs must be in chronological order",
		},
	}
}

// MEVTestCases returns common test cases for MEV validation
func MEVTestCases() []ValidatorTestCase {
	baseConfig := DefaultValidConfig()

	createConfigWithMEV := func(mev *MEVConfig) *EthereumPackageConfig {
		config := &EthereumPackageConfig{
			Participants: baseConfig.Participants,
			MEV:          mev,
		}
		return config
	}

	return []ValidatorTestCase{
		{
			Name: "invalid MEV type",
			Config: createConfigWithMEV(&MEVConfig{
				Type: "invalid",
			}),
			WantErr: "invalid MEV type: invalid",
		},
		{
			Name: "invalid relay URL",
			Config: createConfigWithMEV(&MEVConfig{
				RelayURL: "not-a-url",
			}),
			WantErr: "invalid relay URL: not-a-url",
		},
		{
			Name: "negative max bundle length",
			Config: createConfigWithMEV(&MEVConfig{
				MaxBundleLength: -1,
			}),
			WantErr: "max bundle length cannot be negative",
		},
		{
			Name: "max bundle length too high",
			Config: createConfigWithMEV(&MEVConfig{
				MaxBundleLength: 10001,
			}),
			WantErr: "max bundle length 10001 exceeds maximum of 10000",
		},
	}
}
