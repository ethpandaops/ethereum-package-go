package kurtosis

import (
	"fmt"

	"github.com/ethpandaops/ethereum-package-go/pkg/types"
)

// ConvertServiceInfoToExecutionClient converts Kurtosis ServiceInfo to an ExecutionClient
func ConvertServiceInfoToExecutionClient(service *ServiceInfo, clientType types.ClientType) types.ExecutionClient {
	rpcURL := ""
	wsURL := ""
	engineURL := ""
	metricsURL := ""
	p2pPort := 0

	// Extract URLs from ports
	for portName, portInfo := range service.Ports {
		switch portName {
		case "rpc":
			rpcURL = portInfo.MaybeURL
			if rpcURL == "" && service.IPAddress != "" {
				rpcURL = fmt.Sprintf("http://%s:%d", service.IPAddress, portInfo.Number)
			}
		case "ws":
			wsURL = portInfo.MaybeURL
			if wsURL == "" && service.IPAddress != "" {
				wsURL = fmt.Sprintf("ws://%s:%d", service.IPAddress, portInfo.Number)
			}
		case "engine":
			engineURL = portInfo.MaybeURL
			if engineURL == "" && service.IPAddress != "" {
				engineURL = fmt.Sprintf("http://%s:%d", service.IPAddress, portInfo.Number)
			}
		case "metrics":
			metricsURL = portInfo.MaybeURL
			if metricsURL == "" && service.IPAddress != "" {
				metricsURL = fmt.Sprintf("http://%s:%d", service.IPAddress, portInfo.Number)
			}
		case "p2p":
			p2pPort = int(portInfo.Number)
		}
	}

	// TODO: Extract version and enode from service metadata
	version := "unknown"
	enode := ""

	switch clientType {
	case types.ClientGeth:
		return types.NewGethClient(service.Name, version, rpcURL, wsURL, engineURL, metricsURL, enode, service.Name, service.UUID, p2pPort)
	case types.ClientBesu:
		return types.NewBesuClient(service.Name, version, rpcURL, wsURL, engineURL, metricsURL, enode, service.Name, service.UUID, p2pPort)
	case types.ClientNethermind:
		return types.NewNethermindClient(service.Name, version, rpcURL, wsURL, engineURL, metricsURL, enode, service.Name, service.UUID, p2pPort)
	case types.ClientErigon:
		return types.NewErigonClient(service.Name, version, rpcURL, wsURL, engineURL, metricsURL, enode, service.Name, service.UUID, p2pPort)
	case types.ClientReth:
		return types.NewRethClient(service.Name, version, rpcURL, wsURL, engineURL, metricsURL, enode, service.Name, service.UUID, p2pPort)
	default:
		// Return a generic Geth client as fallback
		return types.NewGethClient(service.Name, version, rpcURL, wsURL, engineURL, metricsURL, enode, service.Name, service.UUID, p2pPort)
	}
}

// ConvertServiceInfoToConsensusClient converts Kurtosis ServiceInfo to a ConsensusClient
func ConvertServiceInfoToConsensusClient(service *ServiceInfo, clientType types.ClientType) types.ConsensusClient {
	beaconAPIURL := ""
	metricsURL := ""
	p2pPort := 0

	// Extract URLs from ports
	for portName, portInfo := range service.Ports {
		switch portName {
		case "beacon", "http":
			beaconAPIURL = portInfo.MaybeURL
			if beaconAPIURL == "" && service.IPAddress != "" {
				beaconAPIURL = fmt.Sprintf("http://%s:%d", service.IPAddress, portInfo.Number)
			}
		case "metrics":
			metricsURL = portInfo.MaybeURL
			if metricsURL == "" && service.IPAddress != "" {
				metricsURL = fmt.Sprintf("http://%s:%d", service.IPAddress, portInfo.Number)
			}
		case "p2p":
			p2pPort = int(portInfo.Number)
		}
	}

	// TODO: Extract version, ENR, and peer ID from service metadata
	version := "unknown"
	enr := ""
	peerID := ""

	switch clientType {
	case types.ClientLighthouse:
		return types.NewLighthouseClient(service.Name, version, beaconAPIURL, metricsURL, enr, peerID, service.Name, service.UUID, p2pPort)
	case types.ClientTeku:
		return types.NewTekuClient(service.Name, version, beaconAPIURL, metricsURL, enr, peerID, service.Name, service.UUID, p2pPort)
	case types.ClientPrysm:
		return types.NewPrysmClient(service.Name, version, beaconAPIURL, metricsURL, enr, peerID, service.Name, service.UUID, p2pPort)
	case types.ClientNimbus:
		return types.NewNimbusClient(service.Name, version, beaconAPIURL, metricsURL, enr, peerID, service.Name, service.UUID, p2pPort)
	case types.ClientLodestar:
		return types.NewLodestarClient(service.Name, version, beaconAPIURL, metricsURL, enr, peerID, service.Name, service.UUID, p2pPort)
	case types.ClientGrandine:
		return types.NewGrandineClient(service.Name, version, beaconAPIURL, metricsURL, enr, peerID, service.Name, service.UUID, p2pPort)
	default:
		// Return a generic Lighthouse client as fallback
		return types.NewLighthouseClient(service.Name, version, beaconAPIURL, metricsURL, enr, peerID, service.Name, service.UUID, p2pPort)
	}
}

// DetectClientType attempts to detect the client type from the service name
func DetectClientType(serviceName string) types.ClientType {
	// Common patterns in ethereum-package service names
	switch {
	case contains(serviceName, "geth"):
		return types.ClientGeth
	case contains(serviceName, "besu"):
		return types.ClientBesu
	case contains(serviceName, "nethermind"):
		return types.ClientNethermind
	case contains(serviceName, "erigon"):
		return types.ClientErigon
	case contains(serviceName, "reth"):
		return types.ClientReth
	case contains(serviceName, "lighthouse"):
		return types.ClientLighthouse
	case contains(serviceName, "teku"):
		return types.ClientTeku
	case contains(serviceName, "prysm"):
		return types.ClientPrysm
	case contains(serviceName, "nimbus"):
		return types.ClientNimbus
	case contains(serviceName, "lodestar"):
		return types.ClientLodestar
	case contains(serviceName, "grandine"):
		return types.ClientGrandine
	default:
		return types.ClientType("unknown")
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsIgnoreCase(s, substr)
}

// containsIgnoreCase checks if string contains substring ignoring case
func containsIgnoreCase(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalIgnoreCase(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

// equalIgnoreCase compares two strings ignoring case
func equalIgnoreCase(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if toLower(a[i]) != toLower(b[i]) {
			return false
		}
	}
	return true
}

// toLower converts a byte to lowercase
func toLower(b byte) byte {
	if 'A' <= b && b <= 'Z' {
		return b + 'a' - 'A'
	}
	return b
}