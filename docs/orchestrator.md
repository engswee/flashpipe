# Flashpipe Orchestrator

The Flashpipe Orchestrator is a high-level command that orchestrates the complete deployment lifecycle for SAP Cloud Integration (CPI) artifacts. It replaces the need for external script wrappers by providing an integrated solution for updating and deploying packages and artifacts.

## Overview

The orchestrator internally calls flashpipe's native functions to:

- **Update packages** - Create or update integration package metadata
- **Update artifacts** - Synchronize artifact content with modified manifests and parameters
- **Deploy artifacts** - Deploy artifacts to runtime and verify deployment status
- **Apply prefixes** - Support multi-tenant/environment scenarios with deployment prefixes
- **Filter processing** - Process only specific packages or artifacts
- **Load configurations** - Support multiple config sources (files, folders, URLs)

Unlike the original CLI wrapper that spawned external processes, the orchestrator uses internal function calls for better performance, error handling, and logging.

## Usage

```bash
flashpipe orchestrator [flags]
```

### Required Flags

Connection details are required (same as other flashpipe commands):

```bash
# OAuth authentication
--tmn-host string
--oauth-host string
--oauth-clientid string
--oauth-clientsecret string

# OR Basic authentication
--tmn-host string
--tmn-userid string
--tmn-password string

# OR use config file (recommended)
--config /path/to/flashpipe.yaml
```

### Authentication via Config File

The orchestrator uses the **standard Flashpipe config file** format, just like all other Flashpipe commands (`deploy`, `update`, etc.).

**Config File Location:**

The orchestrator will automatically look for authentication details in:
1. Path specified with `--config` flag
2. `$HOME/flashpipe.yaml` (auto-detected if exists)
3. Individual command-line flags (if config file not found)

**Config File Format:**

Create a file at `$HOME/flashpipe.yaml` (or any location):

```yaml
# OAuth Authentication (recommended)
tmn-host: tenant.hana.ondemand.com
oauth-host: tenant.authentication.sap.hana.ondemand.com
oauth-clientid: your-client-id
oauth-clientsecret: your-client-secret

# OR Basic Authentication
tmn-host: tenant.hana.ondemand.com
tmn-userid: your-username
tmn-password: your-password
```

**Usage Examples:**

```bash
# Auto-detected from $HOME/flashpipe.yaml
flashpipe orchestrator --update --deploy-config ./deploy-config.yml

# Specify custom config location
flashpipe orchestrator --update \
  --config /path/to/custom-flashpipe.yaml \
  --deploy-config ./deploy-config.yml

# Use individual flags (no config file)
flashpipe orchestrator --update \
  --tmn-host tenant.hana.ondemand.com \
  --oauth-host tenant.authentication.sap.hana.ondemand.com \
  --oauth-clientid your-client-id \
  --oauth-clientsecret your-secret \
  --deploy-config ./deploy-config.yml
```

**Important Notes:**
- The config file is **shared** with all other Flashpipe commands
- If you already use other Flashpipe commands, the orchestrator will use the same config automatically
- Storing credentials in a config file is more secure than passing them as command-line arguments
- The config file uses the same format as standard Flashpipe (not the old standalone CLI format)

### Operation Modes

The orchestrator supports three operation modes:

1. **Update and Deploy** (default) - Updates and deploys artifacts
2. **Update Only** - Only updates artifacts, skips deployment
3. **Deploy Only** - Only deploys artifacts, skips updates

```bash
# Update and deploy (default)
flashpipe orchestrator --update

# Update only, skip deployment
flashpipe orchestrator --update-only

# Deploy only, skip updates
flashpipe orchestrator --deploy-only
```

## Configuration File Format

The orchestrator uses YAML configuration files that define packages and artifacts to process:

```yaml
deploymentPrefix: "DEV"  # Optional prefix for environment isolation

packages:
  - integrationSuiteId: "DeviceManagement"
    packageDir: "DeviceManagement"
    displayName: "Device Management Integration"
    description: "Integration flows for device management"
    sync: true    # Update artifacts (default: true)
    deploy: true  # Deploy artifacts (default: true)
    
    artifacts:
      - artifactId: "MDMDeviceSync"
        artifactDir: "MDMDeviceSync"
        displayName: "MDM Device Synchronization"
        type: "IntegrationFlow"
        sync: true
        deploy: true
        configOverrides:
          SenderURL: "https://qa.example.com/api"
          Timeout: "60000"
      
      - artifactId: "DeviceScripts"
        artifactDir: "DeviceScripts"
        displayName: "Device Helper Scripts"
        type: "ScriptCollection"
        sync: true
        deploy: false  # Don't deploy this artifact

  - integrationSuiteId: "CustomerManagement"
    packageDir: "CustomerManagement"
    displayName: "Customer Management"
    sync: true
    deploy: true
    artifacts:
      - artifactId: "CustomerSync"
        artifactDir: "CustomerSync"
        type: "IntegrationFlow"
```

### Configuration Options

**Package Level:**
- `integrationSuiteId` (required) - Package ID
- `packageDir` (required) - Directory name under packages folder
- `displayName` - Display name for the package
- `description` - Package description
- `short_text` - Short text for package
- `sync` - Whether to update artifacts (default: true)
- `deploy` - Whether to deploy artifacts (default: true)

**Artifact Level:**
- `artifactId` (required) - Artifact ID
- `artifactDir` (required) - Directory name under package folder
- `displayName` - Display name for the artifact
- `type` - Artifact type: IntegrationFlow, ScriptCollection, MessageMapping, ValueMapping
- `sync` - Whether to update this artifact (default: true)
- `deploy` - Whether to deploy this artifact (default: true)
- `configOverrides` - Key-value pairs to override in parameters.prop

## Configuration Sources

The `--deploy-config` flag supports multiple source types:

### Single File

```bash
# Using connection flags
flashpipe orchestrator --update \
  --deploy-config ./001-deploy-config.yml \
  --tmn-host tenant.hana.ondemand.com \
  --oauth-host tenant.authentication.sap.hana.ondemand.com \
  --oauth-clientid your-client-id \
  --oauth-clientsecret your-secret

# Or using config file (recommended)
flashpipe orchestrator --update \
  --config $HOME/flashpipe.yaml \
  --deploy-config ./001-deploy-config.yml
```

### Folder (Multiple Files)

Process all matching config files in a folder (recursively):

```bash
flashpipe orchestrator --update \
  --deploy-config ./configs \
  --config-pattern "*.yml" \
  --tmn-host tenant.hana.ondemand.com \
  --oauth-host tenant.authentication.sap.hana.ondemand.com \
  --oauth-clientid your-client-id \
  --oauth-clientsecret your-secret
```

Files are processed in **alphabetical order**, ensuring deterministic execution.

### Remote URL

Load configuration from a remote URL (e.g., GitHub, internal config server):

```bash
# Public URL
flashpipe orchestrator --update \
  --deploy-config https://raw.githubusercontent.com/org/repo/main/deploy-config.yml \
  --tmn-host tenant.hana.ondemand.com \
  --oauth-host tenant.authentication.sap.hana.ondemand.com \
  --oauth-clientid your-client-id \
  --oauth-clientsecret your-secret

# Protected URL with Bearer token
flashpipe orchestrator --update \
  --deploy-config https://api.example.com/configs/deploy.yml \
  --auth-token "your-bearer-token" \
  --tmn-host tenant.hana.ondemand.com \
  --oauth-host tenant.authentication.sap.hana.ondemand.com \
  --oauth-clientid your-client-id \
  --oauth-clientsecret your-secret

# Protected URL with Basic auth
flashpipe orchestrator --update \
  --deploy-config https://config.example.com/deploy.yml \
  --auth-type basic \
  --username admin \
  --password secret \
  --tmn-host tenant.hana.ondemand.com \
  --oauth-host tenant.authentication.sap.hana.ondemand.com \
  --oauth-clientid your-client-id \
  --oauth-clientsecret your-secret
```

## Merging Multiple Configurations

When loading from a folder with multiple config files, you can choose how to process them:

### Merged Processing (default)

All configs are merged into a single deployment run:

```bash
flashpipe orchestrator --update \
  --deploy-config ./configs \
  --merge-configs=true
```

**Benefits:**
- Single deployment session
- Prefixes from individual configs are applied to package IDs
- Faster overall execution

**Note:** Each config file can have its own `deploymentPrefix`. The CLI `--deployment-prefix` flag is ignored when merging.

### Sequential Processing

Process each config file separately:

```bash
flashpipe orchestrator --update \
  --deploy-config ./configs \
  --merge-configs=false
```

**Benefits:**
- Isolated deployments per config
- Can use CLI `--deployment-prefix` to override each config's prefix
- Errors in one config don't affect others

## Deployment Prefixes

Prefixes support multi-environment deployments (DEV, QA, PROD) from the same codebase:

```bash
# Deploy to DEV environment
flashpipe orchestrator --update \
  --deployment-prefix DEV \
  --deploy-config ./deploy-config.yml

# Deploy to PROD environment
flashpipe orchestrator --update \
  --deployment-prefix PROD \
  --deploy-config ./deploy-config.yml
```

**How Prefixes Work:**

- Package ID: `DeviceManagement` → `DEV_DeviceManagement`
- Package Name: `Device Management` → `DEV - Device Management`
- Artifact ID: `MDMDeviceSync` → `DEV_MDMDeviceSync`

**Validation:**
- Prefixes can only contain: `a-z`, `A-Z`, `0-9`, `_`
- Invalid characters are rejected with an error

## Filtering

Filter which packages or artifacts to process:

### Package Filter

Process only specific packages:

```bash
flashpipe orchestrator --update \
  --package-filter "DeviceManagement,CustomerManagement"
```

### Artifact Filter

Process only specific artifacts:

```bash
flashpipe orchestrator --update \
  --artifact-filter "MDMDeviceSync,CustomerSync"
```

### Combined Filters

```bash
flashpipe orchestrator --update \
  --package-filter "DeviceManagement" \
  --artifact-filter "MDMDeviceSync"
```

Filters work with **OR** logic within each filter type:
- Packages: Process if package ID matches ANY value in package-filter
- Artifacts: Process if artifact ID matches ANY value in artifact-filter

## Directory Structure

The orchestrator expects this directory structure:

```
.
├── packages/
│   ├── DeviceManagement/          # Package directory
│   │   ├── MDMDeviceSync/         # Artifact directory
│   │   │   ├── META-INF/
│   │   │   │   └── MANIFEST.MF
│   │   │   └── src/
│   │   │       └── main/
│   │   │           └── resources/
│   │   │               └── parameters.prop
│   │   └── DeviceScripts/
│   │       └── ...
│   └── CustomerManagement/
│       └── ...
└── 001-deploy-config.yml          # Deployment configuration
```

## Configuration Overrides

The `configOverrides` section in the deployment config allows you to override parameters in `parameters.prop`:

```yaml
artifacts:
  - artifactId: "MDMDeviceSync"
    configOverrides:
      SenderURL: "https://qa.example.com/api"
      Timeout: "60000"
      EnableLogging: "true"
```

**Behavior:**
- Existing parameters are updated with new values
- New parameters are added to the file
- Original file format and line endings are preserved
- Parameters not in overrides remain unchanged

## Advanced Options

### Debug Mode

Enable detailed logging:

```bash
flashpipe orchestrator --update \
  --debug \
  --deploy-config ./deploy-config.yml
```

Shows:
- Config loading details
- File processing steps
- Internal API calls
- Deployment status checks

### Keep Temporary Files

Preserve temporary working directory for troubleshooting:

```bash
flashpipe orchestrator --update \
  --keep-temp \
  --deploy-config ./deploy-config.yml
```

Temporary directory contains:
- Modified MANIFEST.MF files
- Modified parameters.prop files
- Package JSON files
- Artifact working copies

### Custom Packages Directory

Specify a different packages directory:

```bash
flashpipe orchestrator --update \
  --packages-dir ./my-packages \
  --deploy-config ./deploy-config.yml
```

## Examples

### Basic Update and Deploy

```bash
# Using config file (recommended)
flashpipe orchestrator --update \
  --config $HOME/flashpipe.yaml \
  --deploy-config ./001-deploy-config.yml

# Or using connection flags
flashpipe orchestrator --update \
  --deploy-config ./001-deploy-config.yml \
  --tmn-host tenant.hana.ondemand.com \
  --oauth-host tenant.authentication.sap.hana.ondemand.com \
  --oauth-clientid client-id \
  --oauth-clientsecret client-secret
```

### Update Only (No Deployment)

```bash
flashpipe orchestrator --update-only \
  --config $HOME/flashpipe.yaml \
  --deploy-config ./001-deploy-config.yml
```

### Deploy with Environment Prefix

```bash
flashpipe orchestrator --update \
  --config $HOME/flashpipe.yaml \
  --deployment-prefix QA \
  --deploy-config ./deploy-config.yml
```

### Process Multiple Configs from Folder

```bash
flashpipe orchestrator --update \
  --config $HOME/flashpipe.yaml \
  --deploy-config ./configs \
  --config-pattern "deploy-*.yml" \
  --merge-configs=false
```

### Load Config from GitHub

```bash
flashpipe orchestrator --update \
  --config $HOME/flashpipe.yaml \
  --deploy-config https://raw.githubusercontent.com/myorg/configs/main/deploy-dev.yml
```

### Filter Specific Package and Artifacts

```bash
flashpipe orchestrator --update \
  --config $HOME/flashpipe.yaml \
  --package-filter "DeviceManagement" \
  --artifact-filter "MDMDeviceSync,DeviceHelper" \
  --deploy-config ./deploy-config.yml
```

### Reusing Existing Flashpipe Config

If you already use other Flashpipe commands with a config file, the orchestrator will automatically use the same file:

```bash
# If you already have $HOME/flashpipe.yaml set up for other commands
# The orchestrator will use it automatically
flashpipe orchestrator --update --deploy-config ./deploy-config.yml

# This is the same config file used by:
flashpipe deploy --artifact-ids MyFlow  # Uses same config
flashpipe update artifact ...           # Uses same config
```

**Config File Locations (in order of precedence):**
1. Path specified with `--config` flag
2. `$HOME/flashpipe.yaml` (auto-detected)
3. Individual command-line flags

## CI/CD Integration

### GitHub Actions

```yaml
name: Deploy to SAP CPI

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Download Flashpipe
        run: |
          wget https://github.com/engswee/flashpipe/releases/latest/download/flashpipe-linux-amd64
          chmod +x flashpipe-linux-amd64
          mv flashpipe-linux-amd64 /usr/local/bin/flashpipe
      
      - name: Deploy to DEV
        run: |
          flashpipe orchestrator --update \
            --deployment-prefix DEV \
            --deploy-config ./configs/deploy-dev.yml \
            --packages-dir ./packages \
            --tmn-host ${{ secrets.CPI_TMN_HOST }} \
            --oauth-host ${{ secrets.CPI_OAUTH_HOST }} \
            --oauth-clientid ${{ secrets.CPI_CLIENT_ID }} \
            --oauth-clientsecret ${{ secrets.CPI_CLIENT_SECRET }}
```

### Azure DevOps

```yaml
trigger:
  - main

pool:
  vmImage: 'ubuntu-latest'

steps:
- task: Bash@3
  displayName: 'Install Flashpipe'
  inputs:
    targetType: 'inline'
    script: |
      wget https://github.com/engswee/flashpipe/releases/latest/download/flashpipe-linux-amd64
      chmod +x flashpipe-linux-amd64
      sudo mv flashpipe-linux-amd64 /usr/local/bin/flashpipe

- task: Bash@3
  displayName: 'Deploy to QA'
  inputs:
    targetType: 'inline'
    script: |
      flashpipe orchestrator --update \
        --deployment-prefix QA \
        --deploy-config ./001-deploy-config.yml \
        --tmn-host $(CPI_TMN_HOST) \
        --oauth-host $(CPI_OAUTH_HOST) \
        --oauth-clientid $(CPI_CLIENT_ID) \
        --oauth-clientsecret $(CPI_CLIENT_SECRET)
```

### GitLab CI

```yaml
deploy-qa:
  stage: deploy
  image: ubuntu:latest
  before_script:
    - apt-get update && apt-get install -y wget
    - wget https://github.com/engswee/flashpipe/releases/latest/download/flashpipe-linux-amd64
    - chmod +x flashpipe-linux-amd64
    - mv flashpipe-linux-amd64 /usr/local/bin/flashpipe
  script:
    - |
      flashpipe orchestrator --update \
        --deployment-prefix QA \
        --deploy-config ./configs \
        --merge-configs=true \
        --tmn-host $CPI_TMN_HOST \
        --oauth-host $CPI_OAUTH_HOST \
        --oauth-clientid $CPI_CLIENT_ID \
        --oauth-clientsecret $CPI_CLIENT_SECRET
  only:
    - main
```

## Output and Logging

The orchestrator provides detailed progress information:

```
[INFO] Starting flashpipe orchestrator
[INFO] Loading config from: ./001-deploy-config.yml (type: file)
[INFO] Loaded 1 config file(s)
[INFO] Mode: update-and-deploy
[INFO] Packages Directory: ./packages
[INFO] ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[INFO] 📦 Package: DeviceManagement
[INFO] Package ID: DEV_DeviceManagement
[INFO] Package Name: DEV - Device Management Integration
[INFO] Updating package in tenant...
[INFO]   ✓ Package metadata updated
[INFO] Updating artifacts...
[INFO]   Updating: DEV_MDMDeviceSync
[INFO]     ✓ Updated successfully
[INFO] ✓ Updated 1 artifact(s) in package
[INFO] Deploying artifacts...
[INFO]   Deploy: DEV_MDMDeviceSync (type: IntegrationFlow)
[INFO] Deploying artifacts by type...
[INFO] → Deploying 1 artifact(s) of type: IntegrationFlow
[INFO]     ✓ Deployed successfully: DEV_MDMDeviceSync
[INFO] ✓ All artifacts deployed successfully
[INFO] ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[INFO] 📊 DEPLOYMENT SUMMARY
[INFO] ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[INFO] Packages Updated:         1
[INFO] Packages Deployed:        1
[INFO] Packages Failed:          0
[INFO] Packages Filtered:        0
[INFO] ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[INFO] Artifacts Total:          1
[INFO] Artifacts Deployed OK:    1
[INFO] Artifacts Deploy Failed:  0
[INFO] Artifacts Filtered:       0
[INFO] ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[INFO] ✅ Deployment completed successfully
[INFO] ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

## Error Handling

The orchestrator provides clear error messages and continues processing:

- **Package not found**: Warning logged, continues to next package
- **Artifact update failure**: Error logged, artifact skipped for deployment
- **Deployment failure**: Error logged, continues with remaining artifacts
- **Invalid prefix**: Deployment stops with validation error
- **Config load failure**: Stops with error message

Exit codes:
- `0` - Success
- `1` - Failure (check logs for details)

## Performance Considerations

- **Parallel processing**: Not currently implemented (processes sequentially)
- **Batch deployment**: Artifacts are deployed individually for better error tracking
- **Reuse connections**: HTTP client is reused across operations
- **Temporary files**: Cleaned up automatically unless `--keep-temp` is specified

## Migration from External Wrapper

If you were using an external wrapper script that called `flashpipe` as an external command:

**Old approach:**
```bash
#!/bin/bash
flashpipe update package --package-file package.json
flashpipe update artifact --artifact-id MyFlow ...
flashpipe deploy --artifact-ids MyFlow
```

**New approach:**
```bash
flashpipe orchestrator --update --deploy-config ./deploy-config.yml
```

**Benefits:**
- Single process (no subprocess spawning)
- Shared authentication session
- Better error handling and logging
- Config-driven (no script maintenance)
- Built-in filtering and prefixing

## Troubleshooting

### "No config files found"

Check:
- Path is correct: `--deploy-config ./configs`
- Pattern matches files: `--config-pattern "*.yml"`
- Files have correct extension

### "Package directory not found"

Check:
- `packageDir` in config matches actual directory name
- `--packages-dir` points to correct location
- Relative paths are from current working directory

### "Deployment failed"

Enable debug mode:
```bash
flashpipe orchestrator --update --debug
```

Check:
- OAuth credentials are correct
- Tenant host is reachable
- Artifact has no validation errors
- Check CPI tenant logs

### "Duplicate package ID"

When merging configs:
- Each package must have unique ID after prefix is applied
- Use different prefixes or different package IDs

## See Also

- [Partner Directory](./partner-directory.md) - Manage Partner Directory parameters
- [Config Generate](./config-generate.md) - Generate deployment configs from packages
- [FlashPipe Documentation](https://github.com/engswee/flashpipe) - Main flashpipe docs