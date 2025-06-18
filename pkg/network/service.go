package network

import "github.com/ethpandaops/ethereum-package-go/pkg/client"

// Service represents a generic service in the network
type Service struct {
	Name        string
	Type        ServiceType
	ContainerID string
	Ports       []Port
	Status      string
}

// ServiceMetadata contains detailed information about a service
type ServiceMetadata struct {
	Name                string
	ServiceType         ServiceType
	ClientType          client.Type
	Status              string
	ContainerID         string
	IPAddress           string
	Ports               map[string]PortMetadata
	NodeIndex           int
	NodeName            string
	ChainID             uint64
	ValidatorCount      int
	ValidatorStartIndex int
	Version             string
	P2PPort             int
	Enode               string
	ENR                 string
	PeerID              string
}