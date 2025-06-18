package helpers

import (
	"github.com/ethpandaops/ethereum-package-go/pkg/kurtosis"
)

// TestServiceBuilder provides helper methods for creating test service data
type TestServiceBuilder struct{}

// NewTestServiceBuilder creates a new test service builder
func NewTestServiceBuilder() *TestServiceBuilder {
	return &TestServiceBuilder{}
}

// CreateDefaultServices creates a standard set of test services
func (b *TestServiceBuilder) CreateDefaultServices() map[string]*kurtosis.ServiceInfo {
	return map[string]*kurtosis.ServiceInfo{
		"el-1-geth-lighthouse": b.CreateExecutionService("el-1-geth-lighthouse", "uuid-1", "10.0.0.1"),
		"cl-1-lighthouse-geth": b.CreateConsensusService("cl-1-lighthouse-geth", "uuid-2", "10.0.0.2"),
		"apache":               b.CreateApacheService("apache", "uuid-apache", "10.0.0.3"),
		"prometheus":           b.CreatePrometheusService("prometheus", "uuid-prometheus", "10.0.0.4"),
		"grafana":              b.CreateGrafanaService("grafana", "uuid-grafana", "10.0.0.5"),
	}
}

// CreateExecutionService creates a test execution client service
func (b *TestServiceBuilder) CreateExecutionService(name, uuid, ip string) *kurtosis.ServiceInfo {
	return &kurtosis.ServiceInfo{
		Name:      name,
		UUID:      uuid,
		Status:    "running",
		IPAddress: ip,
		Hostname:  name + ".local",
		Ports: map[string]kurtosis.PortInfo{
			"rpc": {
				Number:   8545,
				Protocol: "TCP",
				MaybeURL: "http://" + ip + ":8545",
			},
			"ws": {
				Number:   8546,
				Protocol: "TCP",
				MaybeURL: "ws://" + ip + ":8546",
			},
			"engine": {
				Number:   8551,
				Protocol: "TCP",
				MaybeURL: "http://" + ip + ":8551",
			},
			"metrics": {
				Number:   9090,
				Protocol: "TCP",
				MaybeURL: "http://" + ip + ":9090",
			},
			"p2p": {
				Number:   30303,
				Protocol: "TCP",
			},
		},
	}
}

// CreateConsensusService creates a test consensus client service
func (b *TestServiceBuilder) CreateConsensusService(name, uuid, ip string) *kurtosis.ServiceInfo {
	return &kurtosis.ServiceInfo{
		Name:      name,
		UUID:      uuid,
		Status:    "running",
		IPAddress: ip,
		Hostname:  name + ".local",
		Ports: map[string]kurtosis.PortInfo{
			"http": {
				Number:   5052,
				Protocol: "TCP",
				MaybeURL: "http://" + ip + ":5052",
			},
			"metrics": {
				Number:   5054,
				Protocol: "TCP",
				MaybeURL: "http://" + ip + ":5054",
			},
			"p2p": {
				Number:   9000,
				Protocol: "TCP",
			},
		},
	}
}

// CreateApacheService creates a test Apache service
func (b *TestServiceBuilder) CreateApacheService(name, uuid, ip string) *kurtosis.ServiceInfo {
	return &kurtosis.ServiceInfo{
		Name:      name,
		UUID:      uuid,
		Status:    "running",
		IPAddress: ip,
		Hostname:  name + ".local",
		Ports: map[string]kurtosis.PortInfo{
			"http": {
				Number:   80,
				Protocol: "TCP",
				MaybeURL: "http://" + ip + ":80",
			},
		},
	}
}

// CreatePrometheusService creates a test Prometheus service
func (b *TestServiceBuilder) CreatePrometheusService(name, uuid, ip string) *kurtosis.ServiceInfo {
	return &kurtosis.ServiceInfo{
		Name:      name,
		UUID:      uuid,
		Status:    "running",
		IPAddress: ip,
		Hostname:  name + ".local",
		Ports: map[string]kurtosis.PortInfo{
			"http": {
				Number:   9090,
				Protocol: "TCP",
				MaybeURL: "http://" + ip + ":9090",
			},
		},
	}
}

// CreateGrafanaService creates a test Grafana service
func (b *TestServiceBuilder) CreateGrafanaService(name, uuid, ip string) *kurtosis.ServiceInfo {
	return &kurtosis.ServiceInfo{
		Name:      name,
		UUID:      uuid,
		Status:    "running",
		IPAddress: ip,
		Hostname:  name + ".local",
		Ports: map[string]kurtosis.PortInfo{
			"http": {
				Number:   3000,
				Protocol: "TCP",
				MaybeURL: "http://" + ip + ":3000",
			},
		},
	}
}

// CreateCustomService creates a custom service with specified ports
func (b *TestServiceBuilder) CreateCustomService(name, uuid, ip string, ports map[string]kurtosis.PortInfo) *kurtosis.ServiceInfo {
	return &kurtosis.ServiceInfo{
		Name:      name,
		UUID:      uuid,
		Status:    "running",
		IPAddress: ip,
		Hostname:  name + ".local",
		Ports:     ports,
	}
}
