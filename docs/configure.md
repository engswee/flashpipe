# Configure Command

Configure SAP Cloud Integration artifact parameters using declarative YAML files.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Configuration File Format](#configuration-file-format)
- [Command Reference](#command-reference)
- [Examples](#examples)
- [Multi-Environment Deployments](#multi-environment-deployments)
- [Troubleshooting](#troubleshooting)

---

## Overview

The `configure` command updates configuration parameters for SAP CPI artifacts and optionally deploys them.

**Key Features:**
- Declarative YAML-based configuration
- Batch operations for efficient parameter updates
- Optional deployment after configuration
- Multi-environment support via deployment prefixes
- Dry-run mode to preview changes
- Process single file or folder of configs

**Use Cases:**
- Environment promotion (DEV → QA → PROD)
- Bulk parameter updates
- Configuration as code in CI/CD pipelines
- Disaster recovery

---

## Quick Start

**1. Create config file (`my-config.yml`):**

```yaml
packages:
  - integrationSuiteId: "MyPackage"
    displayName: "My Integration Package"
    
    artifacts:
      - artifactId: "MyFlow"
        displayName: "My Integration Flow"
        type: "Integration"
        version: "active"
        deploy: true
        
        parameters:
          - key: "DatabaseURL"
            value: "jdbc:mysql://localhost:3306/mydb"
          - key: "APIKey"
            value: "${env:API_KEY}"
```

**2. Set environment variables:**

```bash
export API_KEY="your-secret-key"
```

**3. Run command:**

```bash
# Preview changes
flashpipe configure --config-path ./my-config.yml --dry-run

# Apply configuration
flashpipe configure --config-path ./my-config.yml
```

---

## Configuration File Format

### Complete Structure

```yaml
# Optional: Deployment prefix for all packages/artifacts
deploymentPrefix: "DEV_"

packages:
  - integrationSuiteId: "PackageID"        # Required
    displayName: "Package Display Name"     # Required
    deploy: false                           # Optional: deploy all artifacts in package
    
    artifacts:
      - artifactId: "ArtifactID"            # Required
        displayName: "Artifact Name"        # Required
        type: "Integration"                 # Required: Integration|MessageMapping|ScriptCollection|ValueMapping
        version: "active"                   # Optional: default "active"
        deploy: true                        # Optional: deploy this artifact after config
        
        parameters:
          - key: "ParameterName"            # Required
            value: "ParameterValue"         # Required
        
        batch:                              # Optional batch settings
          enabled: true                     # default: true
          batchSize: 90                     # default: 90
```

### Field Reference

#### Package

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `integrationSuiteId` | string | Yes | Package ID in SAP CPI |
| `displayName` | string | Yes | Package display name |
| `deploy` | boolean | No | Deploy all artifacts in package (default: false) |
| `artifacts` | array | Yes | List of artifacts to configure |

#### Artifact

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `artifactId` | string | Yes | Artifact ID in SAP CPI |
| `displayName` | string | Yes | Artifact display name |
| `type` | string | Yes | `Integration`, `MessageMapping`, `ScriptCollection`, or `ValueMapping` |
| `version` | string | No | Version to configure (default: "active") |
| `deploy` | boolean | No | Deploy after configuration (default: false) |
| `parameters` | array | Yes | Configuration parameters |
| `batch` | object | No | Batch processing settings |

#### Parameter

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `key` | string | Yes | Parameter name |
| `value` | string | Yes | Parameter value (supports `${env:VAR}` syntax) |

### Environment Variables

Reference environment variables using `${env:VARIABLE_NAME}`:

```yaml
parameters:
  - key: "DatabasePassword"
    value: "${env:DB_PASSWORD}"
  - key: "OAuthSecret"
    value: "${env:OAUTH_SECRET}"
```

---

## Command Reference

### Syntax

```bash
flashpipe configure [flags]
```

### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--config-path` | `-c` | string | *required* | Path to YAML file or folder |
| `--deployment-prefix` | `-p` | string | `""` | Prefix for package/artifact IDs |
| `--package-filter` | | string | `""` | Filter packages (comma-separated) |
| `--artifact-filter` | | string | `""` | Filter artifacts (comma-separated) |
| `--dry-run` | | bool | `false` | Preview without applying |
| `--deploy-retries` | | int | `5` | Deployment status check retries |
| `--deploy-delay` | | int | `15` | Seconds between deployment checks |
| `--parallel-deployments` | | int | `3` | Max parallel deployments |
| `--batch-size` | | int | `90` | Parameters per batch request |
| `--disable-batch` | | bool | `false` | Disable batch processing |

### Global Configuration (flashpipe.yaml)

```yaml
configure:
  configPath: "./config/dev"
  deploymentPrefix: "DEV_"
  dryRun: false
  deployRetries: 5
  deployDelaySeconds: 15
  parallelDeployments: 3
  batchSize: 90
  disableBatch: false
```

Run without flags:
```bash
flashpipe configure
```

*Note: CLI flags override flashpipe.yaml settings.*

---

## Examples

### Example 1: Basic Configuration

Update parameters without deployment:

```yaml
packages:
  - integrationSuiteId: "CustomerSync"
    displayName: "Customer Synchronization"
    
    artifacts:
      - artifactId: "CustomerDataFlow"
        displayName: "Customer Data Integration"
        type: "Integration"
        deploy: false
        
        parameters:
          - key: "SourceURL"
            value: "https://erp.example.com/api/customers"
          - key: "BatchSize"
            value: "100"
```

```bash
flashpipe configure --config-path ./config.yml
```

### Example 2: Configure and Deploy

Update parameters and deploy:

```yaml
packages:
  - integrationSuiteId: "OrderProcessing"
    displayName: "Order Processing"
    deploy: true
    
    artifacts:
      - artifactId: "OrderValidation"
        type: "Integration"
        deploy: true
        
        parameters:
          - key: "ValidationRules"
            value: "STRICT"
```

```bash
flashpipe configure --config-path ./config.yml
```

### Example 3: Folder-Based

Process all YAML files in a folder:

```
configs/
├── package1.yml
├── package2.yml
└── package3.yml
```

```bash
flashpipe configure --config-path ./configs
```

### Example 4: Filtered Configuration

Configure specific packages or artifacts:

```bash
# Specific packages
flashpipe configure --config-path ./config.yml \
  --package-filter "Package1,Package2"

# Specific artifacts
flashpipe configure --config-path ./config.yml \
  --artifact-filter "Flow1,Flow2"
```

---

## Multi-Environment Deployments

### Strategy 1: Deployment Prefixes

Use same config, different prefixes:

```bash
# Development
flashpipe configure --config-path ./config.yml --deployment-prefix "DEV_"

# QA
flashpipe configure --config-path ./config.yml --deployment-prefix "QA_"

# Production
flashpipe configure --config-path ./config.yml --deployment-prefix "PROD_"
```

### Strategy 2: Separate Folders

Environment-specific configs:

```
config/
├── dev/
│   └── flows.yml
├── qa/
│   └── flows.yml
└── prod/
    └── flows.yml
```

```bash
flashpipe configure --config-path ./config/dev
flashpipe configure --config-path ./config/qa
flashpipe configure --config-path ./config/prod
```

### Strategy 3: Environment Variables

```yaml
parameters:
  - key: "ServiceURL"
    value: "${env:SERVICE_URL}"
  - key: "APIKey"
    value: "${env:API_KEY}"
```

```bash
# Development
export SERVICE_URL="https://dev-api.example.com"
export API_KEY="dev-key"
flashpipe configure --config-path ./config.yml

# Production
export SERVICE_URL="https://api.example.com"
export API_KEY="prod-key"
flashpipe configure --config-path ./config.yml
```

---

## Troubleshooting

### Enable Debug Logging

```bash
export FLASHPIPE_DEBUG=true
flashpipe configure --config-path ./config.yml
```

### Always Use Dry Run First

```bash
flashpipe configure --config-path ./config.yml --dry-run
```

### Common Issues

| Issue | Solution |
|-------|----------|
| Config file not found | Verify path, use absolute path |
| Invalid YAML syntax | Check indentation (spaces not tabs), validate online |
| Authentication failed | Verify credentials in `flashpipe.yaml` |
| Artifact not found | Check ID is correct (case-sensitive), verify prefix |
| Parameter update failed | Try `--disable-batch` flag |
| Deployment timeout | Increase `--deploy-retries` and `--deploy-delay` |
| Environment variable not substituted | Ensure `export` executed before command |

### Summary Output

The command prints detailed statistics:

```
═══════════════════════════════════════════════════════════════════════
CONFIGURATION SUMMARY
═══════════════════════════════════════════════════════════════════════

Configuration Phase:
  Packages processed:       2
  Artifacts processed:      5
  Artifacts configured:     5
  Parameters updated:       23
  
Processing Method:
  Batch requests executed:  3
  Individual requests used: 0

Deployment Phase:
  Deployments successful:   2
  Deployments failed:       0

Overall Status: ✅ SUCCESS
```

---

## Best Practices

✅ **DO:**
- Use `--dry-run` before applying changes
- Version control configuration files
- Use environment variables for secrets
- Test in DEV before promoting to PROD
- Document parameters with comments

❌ **DON'T:**
- Commit secrets to Git
- Skip dry-run in production
- Use hardcoded credentials
- Deploy without testing first

---

## See Also

- [configure-example.yml](../configure-example.yml) - Complete example
- [config-examples/](../config-examples/) - Multi-file examples
- [Orchestrator Command](orchestrator.md) - For full artifact deployments
- [OAuth Setup](oauth_client.md) - Authentication configuration