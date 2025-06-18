package kurtosis

import (
	"fmt"

	"github.com/ethpandaops/ethereum-package-go/pkg/client"
)

// ConvertServiceInfoToExecutionClient converts Kurtosis ServiceInfo to an ExecutionClient
func ConvertServiceInfoToExecutionClient(service *ServiceInfo, clientType client.Type) client.ExecutionClient {
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

	// Extract version and enode from service metadata
	version := extractVersionFromService(service)
	enode := extractEnodeFromService(service)

	return client.NewExecutionClient(clientType, service.Name, version, rpcURL, wsURL, engineURL, metricsURL, enode, service.Name, service.UUID, p2pPort)
}

// ConvertServiceInfoToConsensusClient converts Kurtosis ServiceInfo to a ConsensusClient
func ConvertServiceInfoToConsensusClient(service *ServiceInfo, clientType client.Type) client.ConsensusClient {
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

	// Extract version, ENR, and peer ID from service metadata
	version := extractVersionFromService(service)
	enr := extractENRFromService(service)
	peerID := extractPeerIDFromService(service)

	return client.NewConsensusClient(clientType, service.Name, version, beaconAPIURL, metricsURL, enr, peerID, service.Name, service.UUID, p2pPort)
}

// DetectClientType attempts to detect the client type from the service name
func DetectClientType(serviceName string) client.Type {
	// Common patterns in ethereum-package service names
	// Check consensus clients first as they might have execution client names in them
	// e.g., "cl-1-lighthouse-geth" should be detected as lighthouse, not geth
	
	// First check for consensus client patterns (cl- prefix or consensus keywords)
	if contains(serviceName, "cl-") || contains(serviceName, "consensus") {
		switch {
		case contains(serviceName, "lighthouse"):
			return client.Lighthouse
		case contains(serviceName, "teku"):
			return client.Teku
		case contains(serviceName, "prysm"):
			return client.Prysm
		case contains(serviceName, "nimbus"):
			return client.Nimbus
		case contains(serviceName, "lodestar"):
			return client.Lodestar
		case contains(serviceName, "grandine"):
			return client.Grandine
		}
	}
	
	// Then check for execution client patterns (el- prefix or execution keywords)
	if contains(serviceName, "el-") || contains(serviceName, "execution") {
		switch {
		case contains(serviceName, "geth"):
			return client.Geth
		case contains(serviceName, "besu"):
			return client.Besu
		case contains(serviceName, "nethermind"):
			return client.Nethermind
		case contains(serviceName, "erigon"):
			return client.Erigon
		case contains(serviceName, "reth"):
			return client.Reth
		}
	}
	
	// Fallback to simple matching
	switch {
	case contains(serviceName, "geth"):
		return client.Geth
	case contains(serviceName, "besu"):
		return client.Besu
	case contains(serviceName, "nethermind"):
		return client.Nethermind
	case contains(serviceName, "erigon"):
		return client.Erigon
	case contains(serviceName, "reth"):
		return client.Reth
	case contains(serviceName, "lighthouse"):
		return client.Lighthouse
	case contains(serviceName, "teku"):
		return client.Teku
	case contains(serviceName, "prysm"):
		return client.Prysm
	case contains(serviceName, "nimbus"):
		return client.Nimbus
	case contains(serviceName, "lodestar"):
		return client.Lodestar
	case contains(serviceName, "grandine"):
		return client.Grandine
	default:
		return client.Unknown
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

// extractVersionFromService attempts to extract version from service metadata
func extractVersionFromService(service *ServiceInfo) string {
	// In a real implementation, this would query the service's API
	// or extract from environment variables/labels
	// For now, return a default version based on client type
	clientType := DetectClientType(service.Name)
	switch clientType {
	case client.Geth:
		return "1.13.0"
	case client.Besu:
		return "23.10.0"
	case client.Nethermind:
		return "1.21.0"
	case client.Erigon:
		return "2.54.0"
	case client.Reth:
		return "0.1.0"
	case client.Lighthouse:
		return "4.5.0"
	case client.Teku:
		return "23.11.0"
	case client.Prysm:
		return "4.1.0"
	case client.Nimbus:
		return "23.11.0"
	case client.Lodestar:
		return "1.12.0"
	case client.Grandine:
		return "0.4.0"
	default:
		return "unknown"
	}
}

// extractEnodeFromService attempts to extract enode from service metadata
func extractEnodeFromService(service *ServiceInfo) string {
	// In a real implementation, this would be retrieved from the service
	// For now, construct a placeholder enode
	if service.IPAddress != "" {
		for portName, portInfo := range service.Ports {
			if portName == "p2p" || portName == "tcp" {
				return fmt.Sprintf("enode://0000000000000000000000000000000000000000000000000000000000000000@%s:%d", 
					service.IPAddress, portInfo.Number)
			}
		}
	}
	return ""
}

// extractENRFromService attempts to extract ENR from service metadata
func extractENRFromService(service *ServiceInfo) string {
	// In a real implementation, this would be retrieved from the service's API
	// For now, return empty as ENR needs to be fetched from the beacon node
	return ""
}

// extractPeerIDFromService attempts to extract peer ID from service metadata
func extractPeerIDFromService(service *ServiceInfo) string {
	// In a real implementation, this would be retrieved from the service's API
	// For now, generate a placeholder based on service UUID
	if service.UUID != "" {
		return fmt.Sprintf("16Uiu2HAm%s", service.UUID[:min(10, len(service.UUID))])
	}
	return ""
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}