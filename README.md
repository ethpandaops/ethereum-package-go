# ethereum-package-go

A Go wrapper around ethpandaops/ethereum-package for easy Ethereum devnet creation.

## Overview

ethereum-package-go provides a simple, type-safe Go API for spinning up Ethereum devnets using Kurtosis and the ethereum-package. It follows the familiar testcontainers pattern, making it easy to create full Ethereum networks for testing purposes.

## Features

- **Simple API**: One-line network creation with sensible defaults
- **Type-safe clients**: Strongly-typed interfaces for execution and consensus clients
- **Flexible configuration**: Support for presets, inline configs, and YAML files
- **Comprehensive client support**: All major Ethereum clients (Geth, Besu, Nethermind, Erigon, Reth, Lighthouse, Teku, Prysm, Nimbus, Lodestar, Grandine)
- **Built-in services**: Prometheus, Grafana, Blockscout integration
- **Apache config server**: Access to network configuration files (genesis.ssz, config.yaml, etc.)

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/ethpandaops/ethereum-package-go"
)

func main() {
    ctx := context.Background()
    
    // Start a minimal network (Geth + Lighthouse)
    network, err := ethereum.Run(ctx, ethereum.Minimal())
    if err != nil {
        panic(err)
    }
    defer network.Cleanup(ctx)
    
    // Get execution clients
    gethClients := network.ExecutionClients().Geth()
    if len(gethClients) > 0 {
        fmt.Printf("Geth RPC URL: %s\n", gethClients[0].RPCURL())
    }
    
    // Get consensus clients
    lighthouseClients := network.ConsensusClients().Lighthouse()
    if len(lighthouseClients) > 0 {
        fmt.Printf("Lighthouse Beacon API: %s\n", lighthouseClients[0].BeaconAPIURL())
    }
}
```

## Configuration Options

### Presets

```go
// All execution layer clients with Lighthouse
network, err := ethereum.Run(ctx, ethereum.AllELs())

// All consensus layer clients with Geth  
network, err := ethereum.Run(ctx, ethereum.AllCLs())

// All client combinations (5 EL × 6 CL = 30 nodes)
network, err := ethereum.Run(ctx, ethereum.AllClientsMatrix())
```

### Custom Configuration

```go
network, err := ethereum.Run(ctx,
    ethereum.WithChainID(98765),
    ethereum.WithCustomChain(98765, 6, 16), // chainID, secondsPerSlot, slotsPerEpoch
    ethereum.WithMonitoring(), // Adds Prometheus + Grafana
    ethereum.WithExplorer(),   // Adds Blockscout
    ethereum.WithMEVBoost(),   // Enables MEV-boost
)
```

### Advanced Configuration

```go
config := &types.EthereumPackageConfig{
    Participants: []types.ParticipantConfig{
        {
            ELType: types.ClientGeth,
            CLType: types.ClientLighthouse,
            Count:  3,
            ValidatorCount: 96,
        },
    },
    NetworkParams: &types.NetworkParams{
        ChainID: 54321,
        SecondsPerSlot: 12,
    },
}

network, err := ethereum.Run(ctx, ethereum.WithConfig(config))
```

## Apache Config Server

Access network configuration files through the Apache config server:

```go
apache := network.ApacheConfig()
genesisURL := apache.GenesisSSZURL()      // genesis.ssz
configURL := apache.ConfigYAMLURL()       // config.yaml  
bootnodesURL := apache.BootnodesYAMLURL() // boot_enr.yaml
depositURL := apache.DepositContractBlockURL() // deposit_contract_block.txt
```

## Client Access

```go
// Get all execution clients
for _, client := range network.ExecutionClients().All() {
    fmt.Printf("%s (%s): %s\n", client.Name(), client.Type(), client.RPCURL())
}

// Get specific client types
gethClients := network.ExecutionClients().Geth()
besuClients := network.ExecutionClients().Besu()

// Access client-specific methods
if len(gethClients) > 0 {
    geth := gethClients[0]
    fmt.Printf("Geth RPC: %s\n", geth.RPCURL())
    fmt.Printf("Geth WS: %s\n", geth.WSURL())
    fmt.Printf("Geth Engine: %s\n", geth.EngineURL())
}
```

## Testing Integration

Perfect for integration tests:

```go
func TestMyContract(t *testing.T) {
    ctx := context.Background()
    
    network, err := ethereum.Run(ctx, 
        ethereum.Minimal(),
        ethereum.WithChainID(31337),
    )
    require.NoError(t, err)
    defer network.Cleanup(ctx)
    
    geth := network.ExecutionClients().Geth()[0]
    
    // Use geth.RPCURL() with your web3 client
    // Deploy and test your contracts...
}
```

## Requirements

- Go 1.21+
- Kurtosis engine running locally
- Docker

## Installation

```bash
go get github.com/ethpandaops/ethereum-package-go
```

## Implementation Status

✅ **Completed (Group A - Foundation)**
- Core type system (execution/consensus clients, networks)
- Configuration builder with YAML support
- Presets system (minimal, all ELs, all CLs, matrix)
- Kurtosis client wrapper
- Comprehensive test infrastructure

🚧 **In Progress (Group B - Core API)**
- Main package API with Run function
- Service discovery and type mapping
- Client-specific implementations

📋 **Planned (Groups C-E)**
- Advanced configuration options
- Service accessors and monitoring integration
- Test utilities and helpers
- Full documentation and examples
- CI/CD pipeline

## License

Apache 2.0