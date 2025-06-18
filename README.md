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
    
    // Start a minimal network
    network, err := ethereum.Run(ctx, ethereum.Minimal())
    if err != nil {
        panic(err)
    }
    defer network.Cleanup(ctx)
    
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

## Requirements

- Go 1.21+
- [Kurtosis](https://docs.kurtosis.com/install) running locally
- Docker
