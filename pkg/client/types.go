package client

// Type represents the type of Ethereum client
type Type string

const (
	// Execution clients
	Geth       Type = "geth"
	Besu       Type = "besu"
	Nethermind Type = "nethermind"
	Erigon     Type = "erigon"
	Reth       Type = "reth"
	Ethrex     Type = "ethrex"
	// NimbusEth1 is the nimbus-eth1 execution client. The upstream
	// ethereum-package Kurtosis module accepts this as "nimbus", which
	// collides with the consensus-client name. We expose the disambiguated
	// "nimbusel" alias here and translate on marshal (see MarshalYAML).
	NimbusEth1 Type = "nimbusel"

	// Consensus clients
	Lighthouse Type = "lighthouse"
	Teku       Type = "teku"
	Prysm      Type = "prysm"
	Nimbus     Type = "nimbus"
	Lodestar   Type = "lodestar"
	Grandine   Type = "grandine"

	// Unknown client
	Unknown Type = "unknown"
)

// upstreamELName is the wire name the upstream ethereum-package Kurtosis
// module expects for execution clients whose exported constant differs.
var upstreamELName = map[Type]string{
	NimbusEth1: "nimbus",
}

// IsExecution returns true if the client type is an execution client
func (t Type) IsExecution() bool {
	switch t {
	case Geth, Besu, Nethermind, Erigon, Reth, Ethrex, NimbusEth1:
		return true
	default:
		return false
	}
}

// MarshalYAML translates disambiguated aliases (e.g. NimbusEth1 → "nimbus")
// to the string the upstream ethereum-package Kurtosis module expects. All
// other values round-trip as their raw string.
func (t Type) MarshalYAML() (any, error) {
	if name, ok := upstreamELName[t]; ok {
		return name, nil
	}
	return string(t), nil
}

// IsConsensus returns true if the client type is a consensus client
func (t Type) IsConsensus() bool {
	switch t {
	case Lighthouse, Teku, Prysm, Nimbus, Lodestar, Grandine:
		return true
	default:
		return false
	}
}

// String returns the string representation of the client type
func (t Type) String() string {
	return string(t)
}
