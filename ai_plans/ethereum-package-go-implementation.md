# Ethereum Package Go Implementation Plan

## Executive Summary

The ethereum-package-go project will provide a Go wrapper around the ethpandaops/ethereum-package, enabling Go developers to easily spin up Ethereum devnets for testing purposes. This package addresses a critical gap in the Ethereum development ecosystem where Go developers need programmatic access to full Ethereum networks during their testing cycles. 

Currently, developers must either manually configure complex Ethereum networks or work directly with Kurtosis and ethereum-package's Starlark configuration, which adds significant complexity and learning curve. Our solution will provide a clean, idiomatic Go API that follows the familiar testcontainers pattern, allowing developers to spin up complete Ethereum networks with a few lines of code.

The package will support multiple configuration modes, from simple presets (all ELs, all CLs, or combinations) to advanced custom YAML configurations. All services will be exposed through strongly-typed interfaces, providing compile-time safety and excellent IDE support. The implementation will be thoroughly unit tested and designed for both CI/CD environments and local development workflows.

By abstracting away the complexity of Kurtosis and ethereum-package, this wrapper will significantly reduce the time and effort required to test Ethereum-dependent applications, making it as simple as using testcontainers for traditional databases or services.

## Goals & Objectives

### Primary Goals
- **Simplify Ethereum devnet creation**: Reduce setup time from hours to minutes with a single function call
- **Type-safe service access**: Provide strongly-typed structs for all Ethereum services with compile-time validation
- **Testcontainers-like experience**: Match the familiar API patterns that Go developers already know and love
- **Flexible configuration**: Support both simple presets and advanced custom configurations

### Secondary Objectives
- **Comprehensive test coverage**: Achieve >95% test coverage with unit and integration tests
- **Performance optimization**: Minimize startup time through parallel service initialization
- **Documentation excellence**: Provide clear examples and guides for common use cases
- **CI/CD compatibility**: Ensure seamless integration with GitHub Actions, GitLab CI, and other platforms

## Solution Overview

### Approach
The ethereum-package-go wrapper will act as a bridge between Go applications and the Kurtosis-based ethereum-package. We'll implement a layered architecture with clear separation of concerns: a public API layer following testcontainers patterns, a configuration layer handling YAML generation and validation, a Kurtosis integration layer managing the underlying package execution, and a types layer providing strict typing for all Ethereum services.

The API will prioritize developer experience by offering both simple one-liner setups for common scenarios and advanced configuration options for complex testing needs. We'll use the functional options pattern extensively to maintain API flexibility while keeping the simple cases simple.

### Key Components
1. **Public API Module**: Testcontainers-style Run functions with functional options for easy network creation
2. **Configuration Engine**: YAML builder and validator supporting inline configs, file configs, and presets
3. **Kurtosis Client**: Wrapper around Kurtosis SDK for package execution and service management
4. **Type System**: Strongly-typed structs for all Ethereum clients (EL/CL), validators, and auxiliary services
5. **Service Discovery**: Automatic endpoint detection and typed accessors for all network services
6. **Apache Config Server**: HTTP server providing network configuration endpoints (genesis.ssz, config.yaml, boot_enr.yaml, deposit_contract_block.txt)
7. **Test Utilities**: Helper functions and cleanup mechanisms designed for testing workflows

### Expected Outcomes
- **10x faster test setup**: Developers can spin up full Ethereum networks in seconds instead of manual configuration
- **Zero configuration errors**: Type-safe API prevents common misconfiguration issues at compile time
- **90% code reduction**: Replace hundreds of lines of configuration with a single function call
- **Improved test reliability**: Consistent, reproducible network environments across all test runs
- **Enhanced developer productivity**: Familiar API patterns reduce learning curve to minutes

## Implementation Tasks

### Parallel Execution Groups

#### Group A: Foundation Setup (Execute ALL in parallel)
- [ ] **Task A.1**: Project structure and core types
  - Create pkg/types/execution.go with ExecutionClient interface and implementations (GethClient, BesuClient, etc.)
  - Create pkg/types/execution_test.go with unit tests for execution types
  - Create pkg/types/consensus.go with ConsensusClient interface and implementations
  - Create pkg/types/consensus_test.go with unit tests for consensus types
  - Create pkg/types/network.go with Network, Service, and Port structs
  - Create pkg/types/network_test.go with unit tests for network types
  - Define pkg/types/config.go with configuration types
  - Create pkg/types/config_test.go with validation tests
  - Dependencies: None (can run immediately)
  
- [ ] **Task A.2**: Kurtosis client wrapper
  - Create pkg/kurtosis/client.go with KurtosisClient struct
  - Create pkg/kurtosis/client_test.go with mock Kurtosis API tests
  - Implement RunPackage, GetServices, StopEnclave methods
  - Add pkg/kurtosis/types.go for Kurtosis-specific type conversions
  - Create pkg/kurtosis/types_test.go for conversion tests
  - Create pkg/kurtosis/errors.go for error handling
  - Create pkg/kurtosis/errors_test.go for error scenarios
  - Dependencies: None (can run immediately)
  
- [ ] **Task A.3**: Configuration builder foundation
  - Create pkg/config/builder.go with ConfigBuilder struct
  - Create pkg/config/builder_test.go with builder pattern tests
  - Implement pkg/config/yaml.go for YAML generation
  - Create pkg/config/yaml_test.go with YAML generation/parsing tests
  - Add pkg/config/presets.go with AllELs, AllCLs, AllClientsMatrix presets
  - Create pkg/config/presets_test.go validating all preset configurations
  - Create pkg/config/validator.go for configuration validation
  - Create pkg/config/validator_test.go with validation edge cases
  - Dependencies: None (can run immediately)

- [ ] **Task A.4**: Test infrastructure setup
  - Create test/fixtures/configs/ directory with sample YAMLs
  - Implement test/mocks/kurtosis.go for Kurtosis client mocking
  - Add test/helpers/assertions.go with custom test assertions
  - Create test/integration/setup_test.go for integration test framework
  - Dependencies: None (can run immediately)

#### Group B: Core API Implementation (Execute after Group A completes)
- [ ] **Task B.1**: Main package API with Run function
  - Create ethereum.go with Run function following testcontainers pattern
  - Create ethereum_test.go with integration tests using mock Kurtosis
  - Implement functional options (WithConfig, WithPreset, WithChainID, etc.)
  - Create options_test.go testing all functional options
  - Add network lifecycle management (Start, Stop, Cleanup)
  - Create lifecycle_test.go for lifecycle edge cases
  - Create examples/basic/main.go demonstrating simple usage
  - Dependencies: A.1, A.2, A.3 (needs types, kurtosis client, and config)
  - Can run parallel with: B.2, B.3

- [ ] **Task B.2**: Service discovery and type mapping
  - Create pkg/discovery/mapper.go to map Kurtosis outputs to typed structs
  - Create pkg/discovery/mapper_test.go with comprehensive mapping tests
  - Implement pkg/discovery/endpoints.go for endpoint extraction
  - Create pkg/discovery/endpoints_test.go with URL parsing tests
  - Add pkg/discovery/parser.go for parsing service metadata
  - Create pkg/discovery/parser_test.go with various output format tests
  - Dependencies: A.1, A.2 (needs types and kurtosis client)
  - Can run parallel with: B.1, B.3

- [ ] **Task B.3**: Client-specific implementations
  - Create pkg/clients/execution/ with geth.go, besu.go, nethermind.go, etc.
  - Create pkg/clients/execution/*_test.go for each client implementation
  - Create pkg/clients/consensus/ with lighthouse.go, teku.go, prysm.go, etc.
  - Create pkg/clients/consensus/*_test.go for each client implementation
  - Implement client-specific methods (e.g., GethClient.AdminNodeInfo())
  - Add wait strategies in pkg/clients/wait.go
  - Create pkg/clients/wait_test.go with timeout and retry tests
  - Dependencies: A.1 (needs base types)
  - Can run parallel with: B.1, B.2

#### Group C: Advanced Features (Execute after Group B completes)
- [ ] **Task C.1**: Advanced configuration options
  - Implement WithParticipants for custom participant matrices
  - Add WithMEV for MEV boost configuration
  - Create WithAdditionalServices for auxiliary services
  - Implement WithNetworkParams for custom network parameters
  - Create advanced_options_test.go testing complex configurations
  - Dependencies: B.1 (needs main API structure)
  - Can run parallel with: C.2, C.3

- [ ] **Task C.2**: Service accessors and helpers
  - Create pkg/services/prometheus.go, grafana.go, blockscout.go accessors
  - Create pkg/services/*_test.go for each service accessor
  - Create pkg/services/apache.go for Apache config server accessor
  - Create pkg/services/apache_test.go with config endpoint tests
  - Implement convenience methods like Network.GetPrometheusURL()
  - Add service health checking in pkg/services/health.go
  - Create pkg/services/health_test.go with health check scenarios
  - Create examples/monitoring/main.go showing service usage
  - Dependencies: B.2 (needs service discovery)
  - Can run parallel with: C.1, C.3

- [ ] **Task C.3**: Test utilities and helpers
  - Create pkg/testutil/network.go with test-specific helpers
  - Create pkg/testutil/network_test.go testing the test helpers
  - Implement cleanup tracking and automatic cleanup
  - Add parallel test support utilities
  - Create pkg/testutil/parallel_test.go for concurrent test scenarios
  - Create examples/testing/network_test.go
  - Dependencies: B.1 (needs main API)
  - Can run parallel with: C.1, C.2

#### Group D: Testing and Documentation (Mixed parallel/sequential)
- [ ] **Task D.1**: Unit test suite (Parallel)
  - Write comprehensive unit tests for all packages
  - Achieve >95% code coverage
  - Add table-driven tests for configuration scenarios
  - Implement property-based tests for builders
  - Dependencies: Groups A, B, C must be complete
  - Can run parallel with: D.2

- [ ] **Task D.2**: Integration test suite (Parallel)
  - Create integration tests with real Kurtosis
  - Test all client combinations
  - Verify service discovery accuracy
  - Test cleanup and error scenarios
  - Dependencies: Groups A, B, C must be complete
  - Can run parallel with: D.1

- [ ] **Task D.3**: Documentation and examples (Sequential after D.1, D.2)
  - Write comprehensive README.md
  - Create docs/configuration.md for advanced configs
  - Add docs/api.md with full API reference
  - Create 10+ examples covering common scenarios
  - Dependencies: D.1, D.2 (ensure tests pass first)

- [ ] **Task D.4**: Performance optimization (Sequential after D.3)
  - Profile startup times and optimize bottlenecks
  - Implement connection pooling for Kurtosis client
  - Add caching for repeated configurations
  - Benchmark against manual setup times
  - Dependencies: D.3 (after main functionality is documented)

#### Group E: Release Preparation (Sequential)
- [ ] **Task E.1**: CI/CD pipeline
  - Create .github/workflows/test.yml for automated testing
  - Add .github/workflows/release.yml for releases
  - Configure code coverage reporting
  - Set up automated documentation generation
  - Dependencies: All previous groups complete

- [ ] **Task E.2**: Final polish and release
  - Add CHANGELOG.md and version management
  - Create release automation scripts
  - Perform security audit
  - Publish v1.0.0 with announcement
  - Dependencies: E.1 complete

## Technical Details

### Core API Design
```go
// Basic usage - spin up network with all clients
network, err := ethereum.Run(ctx, 
    ethereum.WithPreset(ethereum.AllClientsMatrix),
)

// Get typed clients
gethClient := network.ExecutionClients().Geth()[0]
rpcURL := gethClient.RPCURL()

// Advanced usage with custom config
network, err := ethereum.Run(ctx,
    ethereum.WithConfigFile("custom-network.yaml"),
    ethereum.WithChainID(12345),
    ethereum.WithAdditionalServices("prometheus", "grafana"),
)

// Access Apache config server endpoints
apache := network.ApacheConfig()
genesisURL := apache.GenesisSSZURL()      // http://127.0.0.1:32966/network-configs/genesis.ssz
configURL := apache.ConfigYAMLURL()        // http://127.0.0.1:32966/network-configs/config.yaml
bootnodesURL := apache.BootnodesYAMLURL() // http://127.0.0.1:32966/network-configs/boot_enr.yaml
depositURL := apache.DepositContractBlockURL() // http://127.0.0.1:32966/network-configs/deposit_contract_block.txt
```

### Type System Example
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

type GethClient struct {
    baseClient
    // Geth-specific fields
}

func (g *GethClient) AdminNodeInfo() (*NodeInfo, error) {
    // Geth-specific admin methods
}

// Apache config server support
type ApacheConfigServer interface {
    URL() string
    GenesisSSZURL() string
    ConfigYAMLURL() string
    BootnodesYAMLURL() string
    DepositContractBlockURL() string
}

type Network interface {
    // ... existing methods ...
    ApacheConfig() ApacheConfigServer
}
```

### Parallel Execution Strategy
- Groups A tasks: 15 minutes if sequential → 5 minutes parallel
- Groups B tasks: 20 minutes if sequential → 10 minutes parallel  
- Groups C tasks: 15 minutes if sequential → 5 minutes parallel
- Total implementation time: 50% reduction through parallelization

## Risk Mitigation
- **Kurtosis API changes**: Version pin Kurtosis SDK, maintain compatibility layer
- **Complex configurations**: Extensive validation and clear error messages
- **Performance concerns**: Benchmark critical paths, optimize service discovery
- **Testing complexity**: Use dependency injection, comprehensive mocking layer