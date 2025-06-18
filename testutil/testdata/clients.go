// Package testdata provides common test data for the ethereum-package-go project.
package testdata

// ConsensusClientData provides test data for consensus clients
var ConsensusClientData = struct {
	DefaultPort      int
	DefaultBeaconURL string
	DefaultMetricsPort int
}{
	DefaultPort:      9000,
	DefaultBeaconURL: "http://localhost:5052",
	DefaultMetricsPort: 5054,
}

// ExecutionClientData provides test data for execution clients
var ExecutionClientData = struct {
	DefaultPort       int
	DefaultRPCPort    int
	DefaultWSPort     int
	DefaultEnginePort int
	DefaultMetricsPort int
}{
	DefaultPort:       30303,
	DefaultRPCPort:    8545,
	DefaultWSPort:     8546,
	DefaultEnginePort: 8551,
	DefaultMetricsPort: 9090,
}

// CommonTestPorts provides commonly used test port numbers
var CommonTestPorts = struct {
	GethRPC         int
	BesuRPC         int
	NethermindRPC   int
	LighthouseBeacon int
	TekuBeacon      int
	PrysmBeacon     int
}{
	GethRPC:         8545,
	BesuRPC:         8545,
	NethermindRPC:   8545,
	LighthouseBeacon: 5052,
	TekuBeacon:      5052,
	PrysmBeacon:     3500,
}

// TestENRs provides test ENR values
var TestENRs = struct {
	Valid   []string
	Invalid []string
}{
	Valid: []string{
		"enr:-abc123",
		"enr:-def456",
		"enr:-ghi789",
		"enr:-jkl012",
		"enr:-mno345",
		"enr:-pqr678",
	},
	Invalid: []string{
		"invalid-enr",
		"",
		"enr:",
		"enr-",
	},
}

// TestPeerIDs provides test peer ID values
var TestPeerIDs = struct {
	Valid   []string
	Invalid []string
}{
	Valid: []string{
		"16Uiu2HAm123",
		"16Uiu2HAm456",
		"16Uiu2HAm789",
		"16Uiu2HAm012",
		"16Uiu2HAm345",
		"16Uiu2HAm678",
	},
	Invalid: []string{
		"invalid-peer-id",
		"",
		"16Uiu2",
		"peer123",
	},
}

// TestEnodes provides test enode values
var TestEnodes = struct {
	Valid   []string
	Invalid []string
}{
	Valid: []string{
		"enode://abc123@127.0.0.1:30303",
		"enode://def456@127.0.0.1:30303",
		"enode://ghi789@127.0.0.1:30303",
		"enode://jkl012@127.0.0.1:30303",
		"enode://mno345@127.0.0.1:30303",
	},
	Invalid: []string{
		"invalid-enode",
		"",
		"enode://",
		"enode://abc123",
		"enode://abc123@",
	},
}