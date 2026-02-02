# Migration Guide: Standalone CLI to Integrated Orchestrator

This guide helps you migrate from the standalone `ci-helper` CLI to the integrated Flashpipe orchestrator command.

## Overview

The standalone CLI has been **fully integrated** into Flashpipe as the `orchestrator` command. All functionality has been ported to use internal Flashpipe functions instead of spawning external processes.

### What Changed

**Before (Standalone CLI):**
- Separate `ci-helper` binary
- Called `flashpipe` binary as external process
- Required both binaries to be installed
- Multiple process spawns for each operation

**After (Integrated Orchestrator):**
- Single `flashpipe` binary
- Uses internal Flashpipe functions directly
- Single process for entire deployment
- Better performance and error handling

## Command Mapping

### Flashpipe Wrapper Command

**Old:**
```bash
ci-helper flashpipe --update \
  --packages-dir ./packages \
  --flashpipe-config ./flashpipe.yml \
  --deploy-config ./001-deploy-config.yml \
  --deployment-prefix DEV
```

**New (Recommended - using --config flag):**
```bash
flashpipe orchestrator --update \
  --config $HOME/flashpipe.yaml \
  --packages-dir ./packages \
  --deploy-config ./001-deploy-config.yml \
  --deployment-prefix DEV
```

**Alternative (using individual flags):**
```bash
flashpipe orchestrator --update \
  --packages-dir ./packages \
  --deploy-config ./001-deploy-config.yml \
  --deployment-prefix DEV \
  --tmn-host tenant.hana.ondemand.com \
  --oauth-host tenant.authentication.sap.hana.ondemand.com \
  --oauth-clientid your-client-id \
  --oauth-clientsecret your-client-secret
```

**Key Changes:**
- Command: `ci-helper flashpipe` → `flashpipe orchestrator`
- Connection details: `--flashpipe-config` → `--config` (standard Flashpipe flag) or individual flags
- The `--config` flag works exactly like other Flashpipe commands
- All other flags remain the same

### Config Generate Command

**Old:**
```bash
ci-helper config --packages-dir ./packages --output ./deploy-config.yml
```

**New:**
```bash
flashpipe config-generate --packages-dir ./packages --output ./deploy-config.yml
```

**Changes:**
- Command: `ci-helper config` → `flashpipe config-generate`
- All flags remain the same

### Partner Directory Commands

**Old:**
```bash
ci-helper pd snapshot --config ./pd-config.yml --output ./partner-directory
ci-helper pd deploy --config ./pd-config.yml --source ./partner-directory
```

**New:**
```bash
flashpipe pd-snapshot --config ./pd-config.yml --output ./partner-directory
flashpipe pd-deploy --config ./pd-config.yml --source ./partner-directory
```

**Changes:**
- Commands: `ci-helper pd snapshot` → `flashpipe pd-snapshot`
- Commands: `ci-helper pd deploy` → `flashpipe pd-deploy`
- All flags remain the same

## Configuration Files

### Deployment Config (No Changes)

The deployment configuration format is **identical**:

```yaml
deploymentPrefix: "DEV"
packages:
  - integrationSuiteId: "DeviceManagement"
    packageDir: "DeviceManagement"
    displayName: "Device Management"
    sync: true
    deploy: true
    artifacts:
      - artifactId: "MDMDeviceSync"
        artifactDir: "MDMDeviceSync"
        type: "IntegrationFlow"
        sync: true
        deploy: true
        configOverrides:
          Timeout: "60000"
```

### Connection Config

The orchestrator uses the **standard Flashpipe config file** format, just like all other Flashpipe commands.

**Old (`flashpipe-config.yml` - standalone CLI format):**
```yaml
host: tenant.hana.ondemand.com
oauth:
  host: tenant.authentication.sap.hana.ondemand.com
  clientid: your-client-id
  clientsecret: your-client-secret
```

**New (`$HOME/flashpipe.yaml` - standard Flashpipe format):**
```yaml
tmn-host: tenant.hana.ondemand.com
oauth-host: tenant.authentication.sap.hana.ondemand.com
oauth-clientid: your-client-id
oauth-clientsecret: your-client-secret
```

**Usage:**
```bash
# Auto-detected from $HOME/flashpipe.yaml
flashpipe orchestrator --update --deploy-config ./deploy-config.yml

# Or specify custom location
flashpipe orchestrator --update \
  --config ./my-flashpipe.yaml \
  --deploy-config ./deploy-config.yml

# Or use individual flags
flashpipe orchestrator --update \
  --tmn-host tenant.hana.ondemand.com \
  --oauth-host tenant.authentication.sap.hana.ondemand.com \
  --oauth-clientid your-client-id \
  --oauth-clientsecret your-client-secret \
  --deploy-config ./deploy-config.yml
```

**Recommendation:** Use the `--config` flag or place the config file at `$HOME/flashpipe.yaml` for automatic detection. This is more secure than passing credentials via command-line flags.

## Flag Changes

### Removed Flags

These flags from the standalone CLI are **no longer needed**:

- `--flashpipe-config` - Use `--config` or individual connection flags
- `--tmn-host`, `--oauth-*` (when using individual flags) - Now use standard Flashpipe flags

### New Flags

These flags are now available:

- `--config` - Path to Flashpipe config file (standard across all commands)
- All standard Flashpipe connection flags

### Renamed Flags

| Old Flag | New Flag | Notes |
|----------|----------|-------|
| None | - | All flags kept the same name |

## Directory Structure

**No changes required** - the directory structure is identical:

```
.
├── packages/
│   ├── DeviceManagement/
│   │   ├── MDMDeviceSync/
│   │   │   ├── META-INF/MANIFEST.MF
│   │   │   └── src/main/resources/parameters.prop
│   │   └── ...
│   └── ...
├── 001-deploy-config.yml
└── flashpipe.yaml  (optional - for connection details)
```

## Step-by-Step Migration

### Step 1: Install Latest Flashpipe

Download the latest Flashpipe release (with orchestrator):

```bash
# Linux/macOS
wget https://github.com/engswee/flashpipe/releases/latest/download/flashpipe-linux-amd64
chmod +x flashpipe-linux-amd64
sudo mv flashpipe-linux-amd64 /usr/local/bin/flashpipe

# Windows
# Download flashpipe-windows-amd64.exe from releases
# Rename to flashpipe.exe
# Add to PATH
```

### Step 2: Update Scripts/CI Pipelines

Replace `ci-helper` commands with `flashpipe` commands:

**Before:**
```bash
ci-helper flashpipe --update \
  --flashpipe-config ./flashpipe.yml \
  --deploy-config ./deploy-config.yml
```

**After (using standard Flashpipe --config flag):**
```bash
flashpipe orchestrator --update \
  --config $HOME/flashpipe.yaml \
  --deploy-config ./deploy-config.yml
```

**Note:** The `--config` flag works exactly like it does for all other Flashpipe commands (`deploy`, `update artifact`, etc.). If you're already using Flashpipe with a config file, the orchestrator will use the same file automatically.

### Step 3: Update Config Files (Optional)

If you used a separate `flashpipe-config.yml`, you can:

**Option A:** Migrate to `$HOME/flashpipe.yaml` (recommended):
```bash
cp flashpipe-config.yml $HOME/flashpipe.yaml
# Edit to use Flashpipe flag naming conventions
```

**Option B:** Use command-line flags:
```bash
flashpipe orchestrator --update \
  --tmn-host tenant.hana.ondemand.com \
  --oauth-host tenant.authentication.sap.hana.ondemand.com \
  --oauth-clientid $CLIENT_ID \
  --oauth-clientsecret $CLIENT_SECRET \
  --deploy-config ./deploy-config.yml
```

### Step 4: Test the Migration

Run a test deployment to a non-production environment:

```bash
flashpipe orchestrator --update-only \
  --deployment-prefix TEST \
  --deploy-config ./deploy-config.yml \
  --packages-dir ./packages \
  --debug
```

Review the logs to ensure everything works as expected.

### Step 5: Remove Standalone CLI

Once migration is complete and tested:

```bash
# Remove the old ci-helper binary
rm /usr/local/bin/ci-helper  # or wherever it was installed

# Remove old config files if migrated
rm ./flashpipe-config.yml  # if you migrated to flashpipe.yaml
```

## CI/CD Pipeline Examples

### GitHub Actions

**Before:**
```yaml
- name: Deploy with ci-helper
  run: |
    ci-helper flashpipe --update \
      --flashpipe-config ./flashpipe.yml \
      --deploy-config ./configs/dev.yml \
      --deployment-prefix DEV
```

**After (recommended - using secrets in config file):**
```yaml
- name: Deploy with Flashpipe Orchestrator
  run: |
    # Create config file from secrets
    echo "tmn-host: ${{ secrets.CPI_TMN_HOST }}" > flashpipe.yaml
    echo "oauth-host: ${{ secrets.CPI_OAUTH_HOST }}" >> flashpipe.yaml
    echo "oauth-clientid: ${{ secrets.CPI_CLIENT_ID }}" >> flashpipe.yaml
    echo "oauth-clientsecret: ${{ secrets.CPI_CLIENT_SECRET }}" >> flashpipe.yaml
    
    flashpipe orchestrator --update \
      --config ./flashpipe.yaml \
      --deploy-config ./configs/dev.yml \
      --deployment-prefix DEV
```

**Alternative (using individual flags):**
```yaml
- name: Deploy with Flashpipe Orchestrator
  run: |
    flashpipe orchestrator --update \
      --deploy-config ./configs/dev.yml \
      --deployment-prefix DEV \
      --tmn-host ${{ secrets.CPI_TMN_HOST }} \
      --oauth-host ${{ secrets.CPI_OAUTH_HOST }} \
      --oauth-clientid ${{ secrets.CPI_CLIENT_ID }} \
      --oauth-clientsecret ${{ secrets.CPI_CLIENT_SECRET }}
```

### Azure DevOps

**Before:**
```yaml
- task: Bash@3
  inputs:
    script: |
      ci-helper flashpipe --update \
        --flashpipe-config ./flashpipe.yml \
        --deploy-config ./deploy-config.yml
```

**After (recommended - using config file):**
```yaml
- task: Bash@3
  inputs:
    script: |
      # Create config file from pipeline variables
      cat > flashpipe.yaml <<EOF
      tmn-host: $(CPI_TMN_HOST)
      oauth-host: $(CPI_OAUTH_HOST)
      oauth-clientid: $(CPI_CLIENT_ID)
      oauth-clientsecret: $(CPI_CLIENT_SECRET)
      EOF
      
      flashpipe orchestrator --update \
        --config ./flashpipe.yaml \
        --deploy-config ./deploy-config.yml
```

**Alternative (using individual flags):**
```yaml
- task: Bash@3
  inputs:
    script: |
      flashpipe orchestrator --update \
        --deploy-config ./deploy-config.yml \
        --tmn-host $(CPI_TMN_HOST) \
        --oauth-host $(CPI_OAUTH_HOST) \
        --oauth-clientid $(CPI_CLIENT_ID) \
        --oauth-clientsecret $(CPI_CLIENT_SECRET)
```

### GitLab CI

**Before:**
```yaml
deploy:
  script:
    - ci-helper flashpipe --update --flashpipe-config ./flashpipe.yml
```

**After (recommended - using config file):**
```yaml
deploy:
  script:
    - |
      cat > flashpipe.yaml <<EOF
      tmn-host: $CPI_TMN_HOST
      oauth-host: $CPI_OAUTH_HOST
      oauth-clientid: $CPI_CLIENT_ID
      oauth-clientsecret: $CPI_CLIENT_SECRET
      EOF
      
      flashpipe orchestrator --update --config ./flashpipe.yaml
```

**Alternative (using individual flags):**
```yaml
deploy:
  script:
    - |
      flashpipe orchestrator --update \
        --tmn-host $CPI_TMN_HOST \
        --oauth-host $CPI_OAUTH_HOST \
        --oauth-clientid $CPI_CLIENT_ID \
        --oauth-clientsecret $CPI_CLIENT_SECRET
```

## Benefits of Migration

### Performance Improvements

1. **Single Process**: No subprocess spawning - ~30-50% faster
2. **Shared Session**: Reuses HTTP client and authentication
3. **Better Memory Usage**: No duplicate process overhead

### Better Error Handling

1. **Detailed Logging**: Integrated with Flashpipe's logging system
2. **Granular Errors**: Know exactly which artifact failed
3. **Contextual Messages**: Better error context and suggestions

### Simplified Deployment

1. **One Binary**: Only need Flashpipe - no additional dependencies
2. **Consistent CLI**: Same flags and behavior across all commands
3. **Better Documentation**: Integrated help and examples

### New Features

1. **Remote Config Loading**: Load configs from URLs
2. **Folder Config Processing**: Process multiple configs at once
3. **Config Merging**: Merge multiple configs into single deployment
4. **Enhanced Filtering**: More flexible package and artifact filtering

## Troubleshooting

### "Command not found: flashpipe"

Ensure Flashpipe is installed and in your PATH:
```bash
which flashpipe
flashpipe --version
```

### "Required flag not set"

The orchestrator requires connection details. Either:
- Use `--config $HOME/flashpipe.yaml`
- Or provide flags: `--tmn-host`, `--oauth-host`, etc.

### "No such file or directory: flashpipe-config.yml"

The old `--flashpipe-config` flag is not supported. The orchestrator uses the **standard Flashpipe `--config` flag** instead:

```bash
# Use standard Flashpipe config flag
flashpipe orchestrator --update --config $HOME/flashpipe.yaml

# Or use individual connection flags
flashpipe orchestrator --update \
  --tmn-host ... --oauth-host ... --oauth-clientid ... --oauth-clientsecret ...
```

The `--config` flag works exactly the same as it does for all other Flashpipe commands.

### "Different output than before"

The orchestrator uses Flashpipe's logging format. Enable `--debug` for detailed output:
```bash
flashpipe orchestrator --update --debug
```

### "Deployment works but slower"

This should not happen - the orchestrator is faster. If you see slower performance:
- Check network connectivity
- Verify no rate limiting on API calls
- Compare with `--debug` output

## Rollback Plan

If you need to rollback:

1. Keep the old `ci-helper` binary during migration period
2. Test thoroughly in non-production environment first
3. Have both binaries available during transition
4. Monitor first production deployment closely

## Getting Help

- **Documentation**: See [docs/orchestrator.md](./docs/orchestrator.md)
- **Issues**: Report issues on the Flashpipe GitHub repository
- **Examples**: Check the examples in the documentation

## FAQ

### Q: Can I use both ci-helper and flashpipe orchestrator?

Yes, during the migration period you can use both. However, we recommend migrating fully to the orchestrator.

### Q: Will my deployment configs still work?

Yes, the deployment config format is 100% compatible. No changes needed.

### Q: Do I need to change my packages structure?

No, the directory structure and package format remain identical.

### Q: Is the orchestrator command stable?

Yes, it uses the same battle-tested Flashpipe internal functions that have been used for years.

### Q: What if I encounter issues?

Enable debug mode (`--debug`) and report issues with the full log output.

### Q: Can I contribute improvements?

Yes! The code is open source. Contributions are welcome.

## Summary

The migration from standalone CLI to integrated orchestrator is straightforward:

1. Replace binary: `ci-helper` → `flashpipe`
2. Update command: `ci-helper flashpipe` → `flashpipe orchestrator`
3. Update config: `--flashpipe-config` → `--config` or individual flags
4. Test thoroughly
5. Deploy with confidence

The orchestrator provides the same functionality with better performance, error handling, and new features - all in a single binary.