# Extra Configuration Example

This example demonstrates how to use the extra configuration features in ethereum-package-go, including:
- Extra mounts for configuration files
- Extra command-line parameters
- Extra environment variables
- Extra container labels

## Features

The ethereum-package-go library now supports all "extra" configuration fields from ethereum-package v6.0.0+:

### Execution Layer (EL)
- `ELExtraParams`: Additional command-line parameters
- `ELExtraMounts`: Mount custom files/directories into containers
- `ELExtraEnvVars`: Set environment variables
- `ELExtraLabels`: Add container labels

### Consensus Layer (CL)
- `CLExtraParams`: Additional command-line parameters
- `CLExtraMounts`: Mount custom files/directories into containers
- `CLExtraEnvVars`: Set environment variables
- `CLExtraLabels`: Add container labels

### Validator Client (VC)
- `VCExtraParams`: Additional command-line parameters
- `VCExtraMounts`: Mount custom files/directories into containers
- `VCExtraEnvVars`: Set environment variables
- `VCExtraLabels`: Add container labels

## Usage

```go
participant := config.NewParticipantBuilder().
    WithCL(client.Prysm).
    WithEL(client.Geth).
    // Mount configuration files
    WithCLExtraMounts(map[string]string{
        "/etc/tysm": "static_files/tysm",
    }).
    // Add command-line parameters
    WithCLExtraParams([]string{
        "--xatu-config-file=/etc/tysm/xatu-config.yaml",
        "--tysm-hook-config-file=/etc/tysm/hooks.yaml",
    }).
    // Set environment variables
    WithCLExtraEnvVars(map[string]string{
        "LOG_LEVEL": "debug",
    }).
    // Add container labels
    WithCLExtraLabels(map[string]string{
        "environment": "production",
    }).
    Build()
```

## Mount Path Requirements

When using extra mounts:
- **Mount paths** (keys) must be absolute paths starting with `/`
- **Source paths** (values) must be relative paths that exist in the ethereum-package repository
- Common source directories in ethereum-package:
  - `static_files/` - Static configuration files
  - `configs/` - Configuration files
  - Files uploaded via Kurtosis

## Example Files

- `main.go` - Full example showing various extra configurations
- `test_extra_config.go` - Simple test to verify YAML generation

## Running the Example

```bash
# Run the full example (requires ethereum-package v6.0.0+)
go run main.go

# Test YAML generation
go run test_extra_config.go
```

## Requirements

- ethereum-package v6.0.0 or later (for mount support)
- Kurtosis CLI installed and running
- Go 1.19+

## Use Cases

Common use cases for extra configurations:

1. **Monitoring Integration**: Mount Xatu or TYSM configuration files for metrics collection
2. **Custom Validator Keys**: Mount validator keys from specific directories
3. **Environment-specific Settings**: Use environment variables for different deployments
4. **Container Organization**: Use labels for better container management
5. **Performance Tuning**: Add extra parameters for optimizing client performance