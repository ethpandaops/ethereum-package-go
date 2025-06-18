package config

import (
	"testing"

	"github.com/ethpandaops/ethereum-package-go/pkg/client"
	"github.com/stretchr/testify/assert"
)

func TestValidatorValidConfig(t *testing.T) {
	config := &EthereumPackageConfig{
		Participants: []ParticipantConfig{
			{
				ELType:         client.Geth,
				CLType:         client.Lighthouse,
				Count:          2,
				ValidatorCount: 64,
			},
		},
		NetworkParams: &NetworkParams{
			Network:                 "kurtosis",
			NetworkID:               "12345",
			SecondsPerSlot:          12,
			NumValidatorKeysPerNode: 64,
			AltairForkEpoch:         0,
			BellatrixForkEpoch:      0,
			CapellaForkEpoch:        10,
			DenebForkEpoch:          20,
			ElectraForkEpoch:        30,
		},
		MEV: &MEVConfig{
			Type:            "full",
			RelayURL:        "http://relay:18550",
			MaxBundleLength: 3,
		},
		AdditionalServices: []AdditionalService{
			{Name: "prometheus"},
			{Name: "grafana"},
		},
		GlobalLogLevel: "info",
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
	tests := ParticipantTestCases()

	// Add additional participant-specific test cases
	additionalTests := []ValidatorTestCase{
		{
			Name: "negative validator count",
			Config: &EthereumPackageConfig{
				Participants: []ParticipantConfig{
					{ELType: client.Geth, CLType: client.Lighthouse, ValidatorCount: -1},
				},
			},
			WantErr: "participant 0: validator count cannot be negative",
		},
		{
			Name: "validator count too high",
			Config: &EthereumPackageConfig{
				Participants: []ParticipantConfig{
					{ELType: client.Geth, CLType: client.Lighthouse, ValidatorCount: 1000001},
				},
			},
			WantErr: "participant 0: validator count cannot exceed 1000000",
		},
	}

	tests = append(tests, additionalTests...)
	RunValidatorTests(t, tests)
}

func TestValidatorNetworkParams(t *testing.T) {
	RunValidatorTests(t, NetworkParamsTestCases())
}

func TestValidatorMEV(t *testing.T) {
	RunValidatorTests(t, MEVTestCases())
}

func TestValidatorAdditionalServices(t *testing.T) {
	tests := []struct {
		name     string
		services []AdditionalService
		wantErr  string
	}{
		{
			name: "missing service name",
			services: []AdditionalService{
				{Name: ""},
			},
			wantErr: "additional service 0: name is required",
		},
		{
			name: "duplicate service",
			services: []AdditionalService{
				{Name: "prometheus"},
				{Name: "prometheus"},
			},
			wantErr: "duplicate additional service: prometheus",
		},
		{
			name: "invalid service name",
			services: []AdditionalService{
				{Name: "invalid-service"},
			},
			wantErr: "invalid additional service name: invalid-service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &EthereumPackageConfig{
				Participants: []ParticipantConfig{
					{ELType: client.Geth, CLType: client.Lighthouse},
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
			wantErr:  "invalid global log level: invalid",
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
			config := &EthereumPackageConfig{
				Participants: []ParticipantConfig{
					{ELType: client.Geth, CLType: client.Lighthouse},
				},
				GlobalLogLevel: tt.logLevel,
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
	// Test execution client validation using ParticipantConfig
	p := ParticipantConfig{ELType: client.Geth, CLType: client.Lighthouse}
	assert.Nil(t, p.Validate(0))

	p.ELType = "invalid"
	assert.NotNil(t, p.Validate(0))

	// Test consensus client validation
	p = ParticipantConfig{ELType: client.Geth, CLType: client.Lighthouse}
	assert.Nil(t, p.Validate(0))

	p.CLType = "invalid"
	assert.NotNil(t, p.Validate(0))

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

	// Test log level validation
	config := &EthereumPackageConfig{
		Participants:   []ParticipantConfig{{ELType: client.Geth, CLType: client.Lighthouse}},
		GlobalLogLevel: "debug",
	}
	assert.Nil(t, config.Validate())

	config.GlobalLogLevel = "invalid"
	assert.NotNil(t, config.Validate())
}
