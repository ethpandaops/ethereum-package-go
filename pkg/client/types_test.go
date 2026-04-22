package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestTypeIsExecution(t *testing.T) {
	tests := []struct {
		clientType  Type
		isExecution bool
	}{
		{Geth, true},
		{Besu, true},
		{Nethermind, true},
		{Erigon, true},
		{Reth, true},
		{Ethrex, true},
		{NimbusEth1, true},
		{Lighthouse, false},
		{Teku, false},
		{Prysm, false},
		{Nimbus, false},
		{Lodestar, false},
		{Grandine, false},
		{Unknown, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.clientType), func(t *testing.T) {
			assert.Equal(t, tt.isExecution, tt.clientType.IsExecution())
		})
	}
}

func TestTypeMarshalYAML(t *testing.T) {
	tests := []struct {
		name   string
		input  Type
		expect string
	}{
		{"geth is passthrough", Geth, "geth\n"},
		{"ethrex is passthrough", Ethrex, "ethrex\n"},
		{"nimbusel translates to upstream nimbus wire name", NimbusEth1, "nimbus\n"},
		{"consensus nimbus is passthrough", Nimbus, "nimbus\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expect, string(data))
		})
	}
}
