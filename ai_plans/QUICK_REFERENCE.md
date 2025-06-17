# ethereum-package-go Quick Reference

## Project Structure
```
ethereum-package-go/
├── ethereum.go                 # Main public API with Run() function
├── options.go                  # Functional options (WithConfig, WithPreset, etc.)
├── pkg/
│   ├── types/                  # Core type definitions
│   │   ├── execution.go        # ExecutionClient interface & structs
│   │   ├── consensus.go        # ConsensusClient interface & structs
│   │   ├── network.go          # Network, Service, Port types
│   │   └── config.go           # Configuration types
│   ├── kurtosis/              # Kurtosis integration
│   │   ├── client.go          # KurtosisClient wrapper
│   │   ├── types.go           # Type conversions
│   │   └── errors.go          # Error handling
│   ├── config/                # Configuration management
│   │   ├── builder.go         # ConfigBuilder
│   │   ├── yaml.go            # YAML generation
│   │   ├── presets.go         # Preset configurations
│   │   └── validator.go       # Config validation
│   ├── discovery/             # Service discovery
│   │   ├── mapper.go          # Map Kurtosis output to types
│   │   ├── endpoints.go       # Extract endpoints
│   │   └── parser.go          # Parse service metadata
│   ├── clients/               # Client implementations
│   │   ├── execution/         # EL clients
│   │   │   ├── geth.go
│   │   │   ├── besu.go
│   │   │   └── ...
│   │   ├── consensus/         # CL clients
│   │   │   ├── lighthouse.go
│   │   │   ├── teku.go
│   │   │   └── ...
│   │   ├── wait.go           # Wait strategies
│   │   └── wait_test.go      # Wait strategy tests
│   ├── services/              # Auxiliary services
│   │   ├── prometheus.go
│   │   ├── prometheus_test.go
│   │   ├── grafana.go
│   │   ├── grafana_test.go
│   │   ├── apache.go          # Apache config server
│   │   ├── apache_test.go
│   │   ├── health.go
│   │   └── health_test.go
│   └── testutil/              # Test utilities
│       ├── network.go
│       ├── network_test.go
│       └── parallel_test.go
├── examples/                   # Usage examples
│   ├── basic/
│   ├── advanced/
│   ├── monitoring/
│   └── testing/
├── test/                      # Test infrastructure
│   ├── fixtures/
│   ├── mocks/
│   ├── helpers/
│   └── integration/
└── docs/                      # Documentation
    ├── configuration.md
    └── api.md
```

## Key API Examples

### Basic Usage

```go
// Simplest case - all clients
network, err := ethereum.Run(ctx, ethereum.WithPreset(ethereum.AllClientsMatrix))
defer network.Cleanup()

// Get services
geth := network.ExecutionClients().Geth()[0]
lighthouse := network.ConsensusClients().Lighthouse()[0]

// Get Apache config endpoints
apache := network.ApacheConfig()
genesisURL := apache.GenesisSSZURL()
```

### With Custom Config
```go
network, err := ethereum.Run(ctx,
    ethereum.WithConfigFile("config.yaml"),
    ethereum.WithChainID(12345),
    ethereum.WithAdditionalServices("prometheus", "grafana"),
)
```

### Inline Config
```go
config := &Config{
    Participants: []Participant{
        {ELType: "geth", CLType: "lighthouse", Count: 2},
        {ELType: "besu", CLType: "teku", Count: 1},
    },
}
network, err := ethereum.Run(ctx, ethereum.WithConfig(config))
```

## Type Interfaces

### ExecutionClient
```go
type ExecutionClient interface {
    Name() string
    Type() ClientType
    RPCURL() string
    WSURL() string  
    EngineURL() string
    Enode() string
    Metrics() string
}
```

### ConsensusClient
```go
type ConsensusClient interface {
    Name() string
    Type() ClientType
    BeaconURL() string
    P2PAddr() string
    ENR() string
    PeerID() string
    Metrics() string
}
```

### Network

```go
type Network interface {
    ExecutionClients() ExecutionClients
    ConsensusClients() ConsensusClients
    ValidatorClients() ValidatorClients
    ChainID() uint64
    NetworkID() uint64
    GenesisTime() time.Time
    Services() map[string]Service
    ApacheConfig() ApacheConfigServer
    Cleanup() error
}
```

### ApacheConfigServer

```go
type ApacheConfigServer interface {
    URL() string
    GenesisSSZURL() string
    ConfigYAMLURL() string
    BootnodesYAMLURL() string
    DepositContractBlockURL() string
}
```

## Presets
- `ethereum.SingleNode` - 1 Geth + 1 Lighthouse
- `ethereum.AllELs` - One of each EL client
- `ethereum.AllCLs` - One of each CL client  
- `ethereum.AllClientsMatrix` - All EL × CL combinations
- `ethereum.MinimalSetup` - 2 nodes for basic testing

## Testing Patterns
```go
func TestMyContract(t *testing.T) {
    network, err := ethereum.Run(context.Background(),
        ethereum.WithPreset(ethereum.SingleNode),
    )
    require.NoError(t, err)
    defer testutil.CleanupNetwork(t, network)
    
    client, err := ethclient.Dial(network.ExecutionClients().Geth()[0].RPCURL())
    require.NoError(t, err)
    
    // Your test code here
}
```