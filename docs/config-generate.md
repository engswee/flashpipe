# Config Generate Command

The `config-generate` command automatically generates or updates deployment configuration files by scanning your packages directory structure.

## Overview

The config generator scans your local packages directory and creates a deployment configuration file (`001-deploy-config.yml`) that can be used with the orchestrator command. It intelligently:

- **Extracts metadata** from package JSON files and artifact MANIFEST.MF files
- **Preserves existing settings** when updating an existing configuration
- **Merges new discoveries** with your existing configuration
- **Filters** by specific packages or artifacts when needed

## Usage

```bash
flashpipe config-generate [flags]
```

### Basic Examples

```bash
# Generate config with defaults (./packages → ./001-deploy-config.yml)
flashpipe config-generate

# Specify custom directories
flashpipe config-generate \
  --packages-dir ./my-packages \
  --output ./my-config.yml

# Generate config for specific packages only
flashpipe config-generate \
  --package-filter "DeviceManagement,OrderProcessing"

# Generate config for specific artifacts only
flashpipe config-generate \
  --artifact-filter "OrderSync,DeviceSync"

# Combine filters
flashpipe config-generate \
  --package-filter "DeviceManagement" \
  --artifact-filter "MDMDeviceSync,DeviceStatusUpdate"
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--packages-dir` | `./packages` | Path to packages directory to scan |
| `--output` | `./001-deploy-config.yml` | Path to output configuration file |
| `--package-filter` | (none) | Comma-separated list of package names to include |
| `--artifact-filter` | (none) | Comma-separated list of artifact names to include |

## How It Works

### 1. Directory Scanning

The generator scans the packages directory with this expected structure:

```
packages/
├── DeviceManagement/
│   ├── DeviceManagement.json          # Package metadata (optional)
│   ├── MDMDeviceSync/
│   │   └── META-INF/MANIFEST.MF       # Artifact metadata
│   └── DeviceStatusUpdate/
│       └── META-INF/MANIFEST.MF
└── OrderProcessing/
    ├── OrderProcessing.json
    └── OrderSync/
        └── META-INF/MANIFEST.MF
```

### 2. Metadata Extraction

**From Package JSON** (e.g., `DeviceManagement.json`):
```json
{
  "Id": "DeviceManagement",
  "Name": "Device Management Integration",
  "Description": "Handles device synchronization",
  "ShortText": "Device Sync"
}
```

**From MANIFEST.MF**:
```
Manifest-Version: 1.0
Bundle-SymbolicName: MDMDeviceSync
Bundle-Name: MDM Device Synchronization
SAP-BundleType: IntegrationFlow
```

Extracts:
- `Bundle-Name` → `displayName`
- `SAP-BundleType` → `type` (e.g., IntegrationFlow, MessageMapping, ScriptCollection)

### 3. Smart Merging

When updating an existing configuration:

**Preserved:**
- ✅ `sync` and `deploy` flags
- ✅ `configOverrides` settings
- ✅ Custom display names and descriptions
- ✅ Deployment prefix

**Added:**
- ✅ Newly discovered packages and artifacts
- ✅ Missing metadata fields

**Removed:**
- ❌ Packages/artifacts no longer in directory (when not using filters)

### 4. Generated Configuration

Example output (`001-deploy-config.yml`):

```yaml
deploymentPrefix: ""
packages:
  - integrationSuiteId: DeviceManagement
    packageDir: DeviceManagement
    displayName: Device Management Integration
    description: Handles device synchronization
    short_text: Device Sync
    sync: true
    deploy: true
    artifacts:
      - artifactId: MDMDeviceSync
        artifactDir: MDMDeviceSync
        displayName: MDM Device Synchronization
        type: IntegrationFlow
        sync: true
        deploy: true
        configOverrides: {}
      - artifactId: DeviceStatusUpdate
        artifactDir: DeviceStatusUpdate
        displayName: Device Status Update Flow
        type: IntegrationFlow
        sync: true
        deploy: true
        configOverrides: {}
```

## Filtering Behavior

### Package Filter

When using `--package-filter`:
- Only specified packages are processed
- Existing packages NOT in the filter are **preserved** in the output
- Statistics show filtered packages separately

```bash
# Only process DeviceManagement, but keep others in config
flashpipe config-generate --package-filter "DeviceManagement"
```

### Artifact Filter

When using `--artifact-filter`:
- Only specified artifacts are processed across all packages
- Existing artifacts NOT in the filter are **preserved** in the output
- Works across package boundaries

```bash
# Only process specific artifacts regardless of package
flashpipe config-generate --artifact-filter "MDMDeviceSync,OrderSync"
```

### Combined Filters

Both filters can be used together:

```bash
# Only process MDMDeviceSync artifact in DeviceManagement package
flashpipe config-generate \
  --package-filter "DeviceManagement" \
  --artifact-filter "MDMDeviceSync"
```

## Statistics Report

After generation, the command displays statistics:

```
Configuration generation completed successfully:

Packages:
  - Preserved: 2
  - Added: 1
  - Filtered: 1
  - Properties extracted: 1
  - Properties preserved: 2

Artifacts:
  - Preserved: 8
  - Added: 2
  - Filtered: 3
  - Display names extracted: 2
  - Display names preserved: 8
  - Types extracted: 2
  - Types preserved: 8

Configuration written to: ./001-deploy-config.yml
```

## Use Cases

### Initial Configuration

Generate a complete configuration from scratch:

```bash
# First time - creates new config
flashpipe config-generate
```

### Update After Changes

After adding new packages or artifacts:

```bash
# Updates existing config, adds new items
flashpipe config-generate
```

### Generate Subset Configuration

Create configuration for a specific subset:

```bash
# Generate config for QA-specific packages
flashpipe config-generate \
  --package-filter "QATestPackage1,QATestPackage2" \
  --output ./qa-deploy-config.yml
```

### Migration/Validation

Regenerate to ensure consistency:

```bash
# Regenerate to validate current structure
flashpipe config-generate --output ./validated-config.yml
```

## Best Practices

1. **Commit Generated Configs**: Add generated files to version control
2. **Review Before Deploying**: Always review generated configs before deployment
3. **Use Filters for Large Projects**: Filter by package/artifact when working with specific components
4. **Preserve Custom Overrides**: The generator never removes your `configOverrides` settings
5. **Regular Updates**: Run after structural changes to your packages directory

## Integration with Orchestrator

The generated configuration is designed to work seamlessly with the orchestrator:

```bash
# Generate configuration
flashpipe config-generate

# Deploy using generated config
flashpipe orchestrator \
  --update \
  --deploy-config ./001-deploy-config.yml \
  --packages-dir ./packages \
  --tmn-host tenant.hana.ondemand.com \
  --oauth-host tenant.authentication.sap.hana.ondemand.com \
  --oauth-clientid your-client-id \
  --oauth-clientsecret your-client-secret
```

## Troubleshooting

### Package Metadata Not Found

If package JSON files don't exist, the generator will still create the configuration but with minimal metadata:

```yaml
- integrationSuiteId: MyPackage
  packageDir: ""
  displayName: ""
  description: ""
  short_text: ""
  sync: true
  deploy: true
```

**Solution**: Create a `{PackageName}.json` file in the package directory.

### Artifact Type Not Detected

If MANIFEST.MF is missing or doesn't have `SAP-BundleType`:

```yaml
- artifactId: MyArtifact
  type: ""
```

**Solution**: Ensure MANIFEST.MF exists and contains `SAP-BundleType` header.

### Existing Config Overwritten

The generator preserves most settings but reorganizes the structure.

**Solution**: Always review the diff before committing changes. Use version control.

### Filter Not Working

Filters are case-sensitive and must match exactly.

**Solution**: Use exact package/artifact names as they appear in the directory structure.

## Related Documentation

- [Orchestrator Command](orchestrator.md) - Deploy using generated configurations
- [Orchestrator YAML Config](orchestrator-yaml-config.md) - Complete orchestrator configuration reference
- [Migration Guide](orchestrator-migration.md) - Migrating from standalone CLI

## Example Workflow

A typical workflow combining config generation and deployment:

```bash
# 1. Sync from SAP CPI to local (if needed)
flashpipe snapshot --sync-package-details

# 2. Generate deployment configuration
flashpipe config-generate

# 3. Review generated configuration
cat ./001-deploy-config.yml

# 4. Deploy using orchestrator
flashpipe orchestrator \
  --update \
  --deploy-config ./001-deploy-config.yml
```

