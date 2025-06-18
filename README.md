# ethereum-package-go

Go wrapper for [ethpandaops/ethereum-package](https://github.com/ethpandaops/ethereum-package) to create Ethereum devnets.

## Installation

```bash
go get github.com/ethpandaops/ethereum-package-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/ethpandaops/ethereum-package-go"
    "github.com/ethpandaops/ethereum-package-go/pkg/types"
)

func main() {
    ctx := context.Background()
    
    // Start a minimal network (auto-cleanup on process exit)
    network, err := ethereum.Run(ctx, ethereum.Minimal())
    if err != nil {
        panic(err)
    }
    // Network auto-cleans up when process exits (testcontainers-style)
    // Optional: defer network.Cleanup(ctx) for explicit cleanup
    
    // Access clients
    clients := network.ExecutionClients().ByType(types.ClientGeth)
    if len(clients) > 0 {
        fmt.Printf("Geth RPC: %s\n", clients[0].RPCURL())
    }
}
```

## Configuration

### Presets

```go
ethereum.Minimal()         // Geth + Lighthouse (1 node)
ethereum.AllELs()          // All execution clients + Lighthouse
ethereum.AllCLs()          // Geth + all consensus clients  
ethereum.AllClientsMatrix() // All combinations (30 nodes)
```

### Custom Options

```go
ethereum.WithChainID(12345)
ethereum.WithCustomChain(12345, 6, 16) // chainID, secondsPerSlot, slotsPerEpoch
ethereum.WithExplorer()                 // Dora
```

### Advanced Config

```go
config := &types.EthereumPackageConfig{
    Participants: []types.ParticipantConfig{{
        ELType: types.ClientGeth,
        CLType: types.ClientLighthouse,
        Count:  3,
    }},
}
network, err := ethereum.Run(ctx, ethereum.WithConfig(config))
```

## Access Clients

```go
// By type
gethClients := network.ExecutionClients().ByType(types.ClientGeth)
lighthouseClients := network.ConsensusClients().ByType(types.ClientLighthouse)

// All clients
for _, client := range network.ExecutionClients().All() {
    fmt.Printf("%s: %s\n", client.Name(), client.RPCURL())
}
```

## Network Configuration

```go
apache := network.ApacheConfig()
genesisURL := apache.GenesisSSZURL()
configURL := apache.ConfigYAMLURL()
```

## Lifecycle Management

### Auto-Cleanup (Default)
Networks automatically clean up when the process exits (testcontainers-style):

```go
// Default behavior - auto cleanup on exit
network, err := ethereum.Run(ctx, ethereum.Minimal())
// Network will be destroyed when process exits
```

### Orphan Networks
Prevent auto-cleanup to create persistent networks:

```go
// Network persists after process exit
network, err := ethereum.Run(ctx,
    ethereum.Minimal(),
    ethereum.WithOrphanOnExit(),
)
// Manual cleanup: kurtosis enclave rm <enclave-name>
```

### Reuse Networks
Connect to existing networks or create reusable ones:

```go
// Reuse existing network or create with specific name
network, err := ethereum.Run(ctx,
    ethereum.Minimal(),
    ethereum.WithReuse("my-persistent-network"),
)

// Or use FindOrCreateNetwork to find existing by name or create new
network, err := ethereum.FindOrCreateNetwork(ctx, "my-persistent-network", ethereum.Minimal())
```

The `FindOrCreateNetwork` function looks for an existing network with the given name and reuses it if found. If no network exists with that name, it creates a new one with the specified configuration.

### Explicit Cleanup
For manual control over cleanup timing:

```go
network, err := ethereum.Run(ctx, ethereum.Minimal())
defer network.Cleanup(ctx) // Explicit cleanup
```

## Requirements

- Go 1.21+
- [Kurtosis](https://docs.kurtosis.com/install) running locally
- Docker
