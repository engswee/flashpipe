# Partner Directory Management

FlashPipe provides comprehensive Partner Directory parameter management for SAP Cloud Integration, allowing you to version control and automate the deployment of Partner Directory parameters alongside your integration artifacts.

## Quick Start

### Prerequisites

- SAP Cloud Integration tenant
- OAuth client credentials with Partner Directory permissions
- FlashPipe installed

### Basic Usage

**1. Download (Snapshot) parameters from SAP CPI:**

```bash
flashpipe pd-snapshot \
  --tmn-host "your-tenant.hana.ondemand.com" \
  --oauth-host "your-tenant.authentication.eu10.hana.ondemand.com" \
  --oauth-clientid "your-client-id" \
  --oauth-clientsecret "your-client-secret"
```

**2. Upload (Deploy) parameters to SAP CPI:**

```bash
flashpipe pd-deploy \
  --tmn-host "your-tenant.hana.ondemand.com" \
  --oauth-host "your-tenant.authentication.eu10.hana.ondemand.com" \
  --oauth-clientid "your-client-id" \
  --oauth-clientsecret "your-client-secret"
```

## Commands

### pd-snapshot

Downloads Partner Directory parameters from SAP CPI to local files.

**Syntax:**
```bash
flashpipe pd-snapshot [flags]
```

**Flags:**
- `--resources-path` - Local directory path (default: `./partner-directory`)
- `--replace` - Overwrite existing local files (default: `true`)
- `--pids` - Filter specific Partner IDs (comma-separated)

**Examples:**

```bash
# Download all parameters
flashpipe pd-snapshot

# Download to custom path
flashpipe pd-snapshot --resources-path "./my-pd-params"

# Download only specific PIDs
flashpipe pd-snapshot --pids "PID_001,PID_002"

# Add-only mode (preserve existing local values)
flashpipe pd-snapshot --replace=false
```

### pd-deploy

Uploads Partner Directory parameters from local files to SAP CPI.

**Syntax:**
```bash
flashpipe pd-deploy [flags]
```

**Flags:**
- `--resources-path` - Local directory path (default: `./partner-directory`)
- `--replace` - Update existing remote parameters (default: `true`)
- `--full-sync` - Delete remote parameters not in local (default: `false`)
- `--dry-run` - Preview changes without executing (default: `false`)
- `--pids` - Filter specific Partner IDs (comma-separated)

**Examples:**

```bash
# Deploy all parameters
flashpipe pd-deploy

# Dry run to preview changes
flashpipe pd-deploy --dry-run

# Deploy only specific PIDs
flashpipe pd-deploy --pids "PID_001,PID_002"

# Add-only mode (don't update existing)
flashpipe pd-deploy --replace=false

# Full sync (delete remote not in local)
flashpipe pd-deploy --full-sync

# Combined: full sync with dry run
flashpipe pd-deploy --full-sync --dry-run
```

## File Structure

Partner Directory parameters are stored in a hierarchical directory structure:

```
partner-directory/
├── PID_001/
│   ├── String.properties
│   └── Binary/
│       ├── certificate.crt
│       ├── config.xml
│       └── _metadata.json
├── PID_002/
│   ├── String.properties
│   └── Binary/
│       └── _metadata.json
└── PID_003/
    └── String.properties
```

### String Parameters

String parameters are stored in `String.properties` files using standard Java properties format:

**Example `String.properties`:**
```properties
API_KEY=abc123def456
ENDPOINT_URL=https://api.example.com/v1
TIMEOUT_SECONDS=30
MULTILINE_VALUE=Line 1\nLine 2\nLine 3
```

**Special Characters:**
- Newlines: `\n`
- Carriage returns: `\r`
- Backslashes: `\\`

### Binary Parameters

Binary parameters are stored as individual files in the `Binary/` subdirectory:

**Example Binary Directory:**
```
Binary/
├── certificate.crt          # Binary parameter named "certificate"
├── transform.xsl            # Binary parameter named "transform"
├── config.xml               # Binary parameter named "config"
└── _metadata.json           # Content type metadata
```

**Metadata File (`_metadata.json`):**
```json
{
  "certificate.crt": "application/x-x509-ca-cert",
  "transform.xsl": "application/xml",
  "config.xml": "application/xml"
}
```

**Supported Content Types:**
- `xml` - XML documents
- `xsl` - XSLT stylesheets
- `xsd` - XML schemas
- `json` - JSON documents
- `txt` - Text files
- `zip` - ZIP archives
- `gz` - Gzip compressed files
- `zlib` - Zlib compressed files
- `crt` - Certificates

## Authentication

### OAuth (Recommended)

Use OAuth 2.0 client credentials flow for secure authentication:

**Using Flags:**
```bash
flashpipe pd-snapshot \
  --tmn-host "tenant.hana.ondemand.com" \
  --oauth-host "tenant.authentication.eu10.hana.ondemand.com" \
  --oauth-clientid "your-client-id" \
  --oauth-clientsecret "your-client-secret"
```

**Using Environment Variables:**
```bash
export FLASHPIPE_TMN_HOST="tenant.hana.ondemand.com"
export FLASHPIPE_OAUTH_HOST="tenant.authentication.eu10.hana.ondemand.com"
export FLASHPIPE_OAUTH_CLIENTID="your-client-id"
export FLASHPIPE_OAUTH_CLIENTSECRET="your-client-secret"

flashpipe pd-snapshot
```

**Using Config File (`~/.flashpipe.yaml`):**
```yaml
tmn-host: "tenant.hana.ondemand.com"
oauth-host: "tenant.authentication.eu10.hana.ondemand.com"
oauth-clientid: "your-client-id"
oauth-clientsecret: "your-client-secret"
```

Then simply run:
```bash
flashpipe pd-snapshot
```

### Basic Authentication (Legacy)

Basic authentication is supported but not recommended:

```bash
flashpipe pd-snapshot \
  --tmn-host "tenant.hana.ondemand.com" \
  --tmn-userid "your-username" \
  --tmn-password "your-password"
```

## Deployment Modes

### Replace Mode (Default)

**Snapshot:** Overwrites existing local files with remote values.
**Deploy:** Updates existing remote parameters with local values.

```bash
# Snapshot with replace
flashpipe pd-snapshot --replace=true

# Deploy with replace
flashpipe pd-deploy --replace=true
```

### Add-Only Mode

**Snapshot:** Only adds new parameters, preserves existing local values.
**Deploy:** Only creates new parameters, skips existing ones.

```bash
# Snapshot add-only
flashpipe pd-snapshot --replace=false

# Deploy add-only
flashpipe pd-deploy --replace=false
```

### Full Sync Mode (Deploy Only)

Ensures remote parameters exactly match local files by deleting remote parameters not present locally.

**⚠️ WARNING:** This will delete remote parameters!

**Important:**
- Only affects PIDs that have local directories
- Parameters in other PIDs are NOT touched
- Recommended to use `--dry-run` first

```bash
# Preview what would be deleted
flashpipe pd-deploy --full-sync --dry-run

# Execute full sync
flashpipe pd-deploy --full-sync
```

## Filtering by Partner ID

Use the `--pids` flag to work with specific Partner IDs:

```bash
# Single PID
flashpipe pd-snapshot --pids "SYSTEM_001"

# Multiple PIDs
flashpipe pd-snapshot --pids "SYSTEM_001,SYSTEM_002,CUSTOMER_API"

# Deploy specific PIDs only
flashpipe pd-deploy --pids "SYSTEM_001,SYSTEM_002"
```

This is useful for:
- Large tenants with many PIDs
- Environment-specific parameters
- Selective deployments

## CI/CD Integration

### Azure Pipelines

```yaml
steps:
  - task: Bash@3
    displayName: 'Deploy Partner Directory'
    env:
      FLASHPIPE_TMN_HOST: $(CPI_HOST)
      FLASHPIPE_OAUTH_HOST: $(CPI_OAUTH_HOST)
      FLASHPIPE_OAUTH_CLIENTID: $(CPI_CLIENT_ID)
      FLASHPIPE_OAUTH_CLIENTSECRET: $(CPI_CLIENT_SECRET)
    inputs:
      targetType: 'inline'
      script: |
        ./flashpipe pd-deploy \
          --resources-path "./partner-directory" \
          --debug
```

### GitHub Actions

```yaml
- name: Deploy Partner Directory
  env:
    FLASHPIPE_TMN_HOST: ${{ secrets.CPI_HOST }}
    FLASHPIPE_OAUTH_HOST: ${{ secrets.CPI_OAUTH_HOST }}
    FLASHPIPE_OAUTH_CLIENTID: ${{ secrets.CPI_CLIENT_ID }}
    FLASHPIPE_OAUTH_CLIENTSECRET: ${{ secrets.CPI_CLIENT_SECRET }}
  run: |
    ./flashpipe pd-deploy \
      --resources-path "./partner-directory" \
      --debug
```

### GitLab CI

```yaml
deploy-partner-directory:
  script:
    - |
      ./flashpipe pd-deploy \
        --tmn-host "${CPI_HOST}" \
        --oauth-host "${CPI_OAUTH_HOST}" \
        --oauth-clientid "${CPI_CLIENT_ID}" \
        --oauth-clientsecret "${CPI_CLIENT_SECRET}" \
        --resources-path "./partner-directory" \
        --debug
```

## Best Practices

### 1. Version Control

Always commit your `partner-directory` folder to Git:

```bash
git add partner-directory/
git commit -m "Update Partner Directory parameters for PID_001"
git push
```

### 2. Use Dry Run Before Deploy

Preview changes before executing:

```bash
flashpipe pd-deploy --dry-run
```

### 3. Enable Debug Logging

Use `--debug` for troubleshooting:

```bash
flashpipe pd-deploy --debug
```

### 4. Secure Credentials

**DO:**
- Use environment variables in CI/CD
- Store credentials in config file with restricted permissions
- Use OAuth instead of Basic Auth

**DON'T:**
- Hardcode credentials in scripts
- Commit credentials to version control
- Share credentials in plain text

### 5. Test in Non-Production First

```bash
# Development tenant
flashpipe pd-deploy \
  --tmn-host "dev-tenant.hana.ondemand.com" \
  --resources-path "./partner-directory"

# After testing, deploy to production
flashpipe pd-deploy \
  --tmn-host "prod-tenant.hana.ondemand.com" \
  --resources-path "./partner-directory"
```

### 6. Use PID Naming Conventions

Organize PIDs by purpose:
- `DEV_*` - Development systems
- `TEST_*` - Test systems  
- `PROD_*` - Production systems
- `API_*` - External API configurations

## Workflow Examples

### Initial Setup

```bash
# 1. Snapshot current state
flashpipe pd-snapshot --resources-path "./partner-directory"

# 2. Add to version control
git add partner-directory/
git commit -m "Initial Partner Directory snapshot"
git push
```

### Update Parameters

```bash
# 1. Modify local files
nano partner-directory/PID_001/String.properties

# 2. Preview changes
flashpipe pd-deploy --dry-run

# 3. Deploy
flashpipe pd-deploy

# 4. Commit changes
git add partner-directory/
git commit -m "Update API_KEY for PID_001"
git push
```

### Migrate Between Tenants

```bash
# 1. Snapshot from source tenant
flashpipe pd-snapshot \
  --tmn-host "source-tenant.hana.ondemand.com" \
  --resources-path "./pd-export"

# 2. Deploy to target tenant
flashpipe pd-deploy \
  --tmn-host "target-tenant.hana.ondemand.com" \
  --resources-path "./pd-export"
```

### Environment Promotion

```bash
# 1. Snapshot from DEV
flashpipe pd-snapshot \
  --tmn-host "dev-tenant.hana.ondemand.com" \
  --pids "DEV_SYSTEM_001" \
  --resources-path "./dev-params"

# 2. Copy and modify for TEST
cp -r ./dev-params/DEV_SYSTEM_001 ./test-params/TEST_SYSTEM_001
# Edit test-params/TEST_SYSTEM_001/String.properties as needed

# 3. Deploy to TEST
flashpipe pd-deploy \
  --tmn-host "test-tenant.hana.ondemand.com" \
  --resources-path "./test-params"
```

## Config File Reference

### Connection Settings (Top Level)

```yaml
# OAuth Authentication (recommended)
tmn-host: tenant.hana.ondemand.com
oauth-host: tenant.authentication.sap.hana.ondemand.com
oauth-clientid: ${OAUTH_CLIENT_ID}      # Can use env vars
oauth-clientsecret: ${OAUTH_CLIENT_SECRET}

# OR Basic Authentication
tmn-host: tenant.hana.ondemand.com
tmn-userid: your-username
tmn-password: ${PASSWORD}                # Use env vars for secrets
```

### Partner Directory Snapshot Settings

```yaml
pd-snapshot:
  resources-path: ./partner-directory   # Where to save files
  replace: true                          # Replace existing files
  pids:                                  # Optional: filter PIDs
    - SAP_SYSTEM_001
    - CUSTOMER_API
```

### Partner Directory Deploy Settings

```yaml
pd-deploy:
  resources-path: ./partner-directory   # Where to read files from
  replace: true                          # Replace existing values in CPI
  full-sync: true                        # Delete remote params not in local
  dry-run: false                         # Preview changes without applying
  pids:                                  # Optional: filter PIDs
    - SAP_SYSTEM_001
    - CUSTOMER_API
```

### Complete Example

**flashpipe-cpars-prod.yml:**
```yaml
# Production Partner Directory Configuration
tmn-host: prod-tenant.hana.ondemand.com
oauth-host: prod-tenant.authentication.sap.hana.ondemand.com
oauth-clientid: ${PROD_OAUTH_CLIENT_ID}
oauth-clientsecret: ${PROD_OAUTH_CLIENT_SECRET}

pd-snapshot:
  resources-path: ./cpars-prod
  replace: true

pd-deploy:
  resources-path: ./cpars-prod
  replace: true
  full-sync: true
  dry-run: false
  pids:
    - PROD_SAP_SYSTEM
    - PROD_PARTNER_API
```

**Usage:**
```bash
# Set secrets via environment
export PROD_OAUTH_CLIENT_ID="your-client-id"
export PROD_OAUTH_CLIENT_SECRET="your-client-secret"

# Snapshot
flashpipe pd-snapshot --config flashpipe-cpars-prod.yml

# Deploy with dry-run first
flashpipe pd-deploy --config flashpipe-cpars-prod.yml --dry-run

# Deploy for real
flashpipe pd-deploy --config flashpipe-cpars-prod.yml
```

### Environment-Specific Configurations

**flashpipe-cpars-dev.yml:**
```yaml
tmn-host: dev-tenant.hana.ondemand.com
oauth-host: dev-tenant.authentication.sap.hana.ondemand.com
oauth-clientid: ${DEV_OAUTH_CLIENT_ID}
oauth-clientsecret: ${DEV_OAUTH_CLIENT_SECRET}

pd-deploy:
  resources-path: ./cpars
  replace: true
  full-sync: false  # Safer for dev - don't auto-delete
  dry-run: false
```

**flashpipe-cpars-qa.yml:**
```yaml
tmn-host: qa-tenant.hana.ondemand.com
oauth-host: qa-tenant.authentication.sap.hana.ondemand.com
oauth-clientid: ${QA_OAUTH_CLIENT_ID}
oauth-clientsecret: ${QA_OAUTH_CLIENT_SECRET}

pd-deploy:
  resources-path: ./cpars
  replace: true
  full-sync: true   # QA can mirror prod config
  dry-run: false
  pids:
    - QA_SAP_SYSTEM
    - QA_PARTNER_API
```

**CI/CD Usage:**
```bash
# Deploy to different environments using different configs
flashpipe pd-deploy --config flashpipe-cpars-dev.yml
flashpipe pd-deploy --config flashpipe-cpars-qa.yml
flashpipe pd-deploy --config flashpipe-cpars-prod.yml

# Override specific settings via command-line
flashpipe pd-deploy --config flashpipe-cpars-qa.yml --dry-run
flashpipe pd-deploy --config flashpipe-cpars-prod.yml --pids "CRITICAL_PARTNER"
```

### Benefits of Config Files

1. **Environment Separation** - One config per environment
2. **Version Control** - Track changes to deployment settings
3. **Simplified Commands** - No need to remember all flags
4. **Secret Management** - Use environment variables for credentials
5. **Team Sharing** - Easy to share standard configurations
6. **CI/CD Ready** - Simple integration with pipelines

## Troubleshooting

### Authentication Failed

**Error:** `failed to get OAuth token`

**Solution:**
- Verify OAuth host is correct (no `https://` prefix)
- Check client credentials have Partner Directory permissions
- Ensure OAuth path is `/oauth/token`

### Permission Denied

**Error:** `403 Forbidden`

**Solution:**
- Verify OAuth client has Partner Directory API permissions
- Check user roles in SAP BTP cockpit
- Ensure process integration role is assigned

### Parameter Not Found

**Error:** `404 Not Found`

**Solution:**
- Check PID and parameter ID exist on tenant
- Verify spelling and case sensitivity
- Use snapshot to see existing parameters

### Batch Operation Failed

**Error:** `batch request failed with status 500`

**Solution:**
- Reduce batch size (default is 90)
- Check individual parameter errors in debug logs
- Enable `--debug` for detailed error messages

### File Encoding Issues

**Error:** Binary parameter corrupted after upload

**Solution:**
- Ensure binary files are not modified by text editors
- Check file extension matches content type in metadata
- Verify base64 encoding/decoding is working

## Performance

### Batch Processing

FlashPipe uses OData $batch requests for efficient bulk operations:
- Default batch size: 90 operations per request
- Automatic batching for create, update, and delete operations
- Single API call for multiple parameters

### Optimization Tips

1. **Use filtering for large tenants:**
   ```bash
   flashpipe pd-deploy --pids "SPECIFIC_PID"
   ```

2. **Deploy only changed PIDs:**
   ```bash
   # Determine which PIDs changed in Git
   CHANGED_PIDS=$(git diff --name-only HEAD~1 partner-directory/ | cut -d/ -f2 | sort -u | tr '\n' ',')
   flashpipe pd-deploy --pids "${CHANGED_PIDS}"
   ```

3. **Parallel tenant deployment:**
   ```bash
   # Deploy to multiple tenants in parallel
   flashpipe pd-deploy --tmn-host "tenant1.hana.ondemand.com" &
   flashpipe pd-deploy --tmn-host "tenant2.hana.ondemand.com" &
   wait
   ```

## Reference

### Global Flags

All FlashPipe global flags are supported:

- `--config` - Config file path (default: `~/.flashpipe.yaml`)
- `--debug` - Enable debug logging
- `--tmn-host` - Tenant management node host
- `--oauth-host` - OAuth token server host
- `--oauth-clientid` - OAuth client ID
- `--oauth-clientsecret` - OAuth client secret
- `--oauth-path` - OAuth token path (default: `/oauth/token`)
- `--tmn-userid` - Basic auth user ID
- `--tmn-password` - Basic auth password

### Exit Codes

- `0` - Success
- `1` - Error occurred

### Log Levels

- `INFO` - Normal operation messages
- `DEBUG` - Detailed operation tracking (use `--debug`)
- `WARN` - Non-fatal errors and warnings
- `ERROR` - Fatal errors

## Support

For issues, questions, or contributions:
- Review the main [FlashPipe documentation](../README.md)
- Check [PARTNER_DIRECTORY_MIGRATION.md](../PARTNER_DIRECTORY_MIGRATION.md) for technical details
- Enable `--debug` for detailed logs
- Check SAP CPI tenant connectivity and permissions