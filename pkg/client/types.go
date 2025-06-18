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

// IsExecution returns true if the client type is an execution client
func (t Type) IsExecution() bool {
	switch t {
	case Geth, Besu, Nethermind, Erigon, Reth:
		return true
	default:
		return false
	}
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
