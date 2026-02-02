# Partner Directory Config File Examples

Quick reference for Partner Directory configuration files.

## Minimal Config

**flashpipe-pd.yml:**
```yaml
tmn-host: tenant.hana.ondemand.com
oauth-host: tenant.authentication.sap.hana.ondemand.com
oauth-clientid: your-client-id
oauth-clientsecret: your-client-secret
```

**Usage:**
```bash
flashpipe pd-snapshot --config flashpipe-pd.yml
flashpipe pd-deploy --config flashpipe-pd.yml
```

## Full Config with All Options

**flashpipe-pd-full.yml:**
```yaml
# Connection Settings
tmn-host: tenant.hana.ondemand.com
oauth-host: tenant.authentication.sap.hana.ondemand.com
oauth-clientid: ${OAUTH_CLIENT_ID}
oauth-clientsecret: ${OAUTH_CLIENT_SECRET}

# Snapshot Settings
pd-snapshot:
  resources-path: ./partner-directory
  replace: true
  pids:
    - SAP_SYSTEM_001
    - CUSTOMER_API

# Deploy Settings  
pd-deploy:
  resources-path: ./partner-directory
  replace: true
  full-sync: true
  dry-run: false
  pids:
    - SAP_SYSTEM_001
    - CUSTOMER_API
```

## Environment-Specific Configs

### Development Environment

**flashpipe-pd-dev.yml:**
```yaml
tmn-host: dev-tenant.hana.ondemand.com
oauth-host: dev-tenant.authentication.sap.hana.ondemand.com
oauth-clientid: ${DEV_CLIENT_ID}
oauth-clientsecret: ${DEV_CLIENT_SECRET}

pd-snapshot:
  resources-path: ./cpars-dev
  replace: true

pd-deploy:
  resources-path: ./cpars-dev
  replace: true
  full-sync: false  # Don't auto-delete in dev
  dry-run: false
```

### QA Environment

**flashpipe-pd-qa.yml:**
```yaml
tmn-host: qa-tenant.hana.ondemand.com
oauth-host: qa-tenant.authentication.sap.hana.ondemand.com
oauth-clientid: ${QA_CLIENT_ID}
oauth-clientsecret: ${QA_CLIENT_SECRET}

pd-snapshot:
  resources-path: ./cpars-qa
  replace: true

pd-deploy:
  resources-path: ./cpars-qa
  replace: true
  full-sync: true   # QA can use full sync
  dry-run: false
```

### Production Environment

**flashpipe-pd-prod.yml:**
```yaml
tmn-host: prod-tenant.hana.ondemand.com
oauth-host: prod-tenant.authentication.sap.hana.ondemand.com
oauth-clientid: ${PROD_CLIENT_ID}
oauth-clientsecret: ${PROD_CLIENT_SECRET}

pd-snapshot:
  resources-path: ./cpars-prod
  replace: true

pd-deploy:
  resources-path: ./cpars-prod
  replace: true
  full-sync: true
  dry-run: false
  pids:  # Only specific PIDs in prod
    - PROD_SAP_SYSTEM
    - PROD_PARTNER_001
    - PROD_PARTNER_002
```

## Use Case Examples

### Safe Production Deploy (Dry Run First)

**flashpipe-pd-prod-safe.yml:**
```yaml
tmn-host: prod-tenant.hana.ondemand.com
oauth-host: prod-tenant.authentication.sap.hana.ondemand.com
oauth-clientid: ${PROD_CLIENT_ID}
oauth-clientsecret: ${PROD_CLIENT_SECRET}

pd-deploy:
  resources-path: ./cpars-prod
  replace: true
  full-sync: true
  dry-run: true  # Always dry-run by default
```

**Usage:**
```bash
# Review changes first
flashpipe pd-deploy --config flashpipe-pd-prod-safe.yml

# If OK, override dry-run
flashpipe pd-deploy --config flashpipe-pd-prod-safe.yml --dry-run=false
```

### Multi-Region Setup

**flashpipe-pd-eu.yml:**
```yaml
tmn-host: eu-tenant.hana.ondemand.com
oauth-host: eu-tenant.authentication.eu10.hana.ondemand.com
oauth-clientid: ${EU_CLIENT_ID}
oauth-clientsecret: ${EU_CLIENT_SECRET}

pd-deploy:
  resources-path: ./cpars
  replace: true
  full-sync: true
```

**flashpipe-pd-us.yml:**
```yaml
tmn-host: us-tenant.hana.ondemand.com
oauth-host: us-tenant.authentication.us10.hana.ondemand.com
oauth-clientid: ${US_CLIENT_ID}
oauth-clientsecret: ${US_CLIENT_SECRET}

pd-deploy:
  resources-path: ./cpars
  replace: true
  full-sync: true
```

**Usage:**
```bash
# Deploy to both regions
flashpipe pd-deploy --config flashpipe-pd-eu.yml
flashpipe pd-deploy --config flashpipe-pd-us.yml
```

### Incremental Updates (No Replace)

**flashpipe-pd-incremental.yml:**
```yaml
tmn-host: tenant.hana.ondemand.com
oauth-host: tenant.authentication.sap.hana.ondemand.com
oauth-clientid: ${CLIENT_ID}
oauth-clientsecret: ${CLIENT_SECRET}

pd-deploy:
  resources-path: ./cpars
  replace: false  # Only add new, don't update existing
  full-sync: false
  dry-run: false
```

### Specific PID Management

**flashpipe-pd-single-partner.yml:**
```yaml
tmn-host: tenant.hana.ondemand.com
oauth-host: tenant.authentication.sap.hana.ondemand.com
oauth-clientid: ${CLIENT_ID}
oauth-clientsecret: ${CLIENT_SECRET}

pd-snapshot:
  resources-path: ./partner-specific
  pids:
    - PARTNER_ABC_123

pd-deploy:
  resources-path: ./partner-specific
  replace: true
  full-sync: false
  pids:
    - PARTNER_ABC_123
```

## CI/CD Pipeline Examples

### GitHub Actions

**.github/workflows/deploy-pd.yml:**
```yaml
name: Deploy Partner Directory

on:
  push:
    branches: [main]
    paths:
      - 'cpars/**'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Deploy to QA
        env:
          QA_CLIENT_ID: ${{ secrets.QA_OAUTH_CLIENT_ID }}
          QA_CLIENT_SECRET: ${{ secrets.QA_OAUTH_CLIENT_SECRET }}
        run: |
          flashpipe pd-deploy --config flashpipe-pd-qa.yml
      
      - name: Deploy to Production
        if: github.ref == 'refs/heads/main'
        env:
          PROD_CLIENT_ID: ${{ secrets.PROD_OAUTH_CLIENT_ID }}
          PROD_CLIENT_SECRET: ${{ secrets.PROD_OAUTH_CLIENT_SECRET }}
        run: |
          flashpipe pd-deploy --config flashpipe-pd-prod.yml --dry-run
          # Manual approval required for actual deploy
```

### Azure DevOps

**azure-pipelines.yml:**
```yaml
trigger:
  branches:
    include:
      - main
  paths:
    include:
      - cpars/*

pool:
  vmImage: 'ubuntu-latest'

steps:
- task: Bash@3
  displayName: 'Deploy to QA'
  env:
    QA_CLIENT_ID: $(QA_OAUTH_CLIENT_ID)
    QA_CLIENT_SECRET: $(QA_OAUTH_CLIENT_SECRET)
  inputs:
    script: |
      flashpipe pd-deploy --config flashpipe-pd-qa.yml

- task: Bash@3
  displayName: 'Deploy to Production (Dry Run)'
  env:
    PROD_CLIENT_ID: $(PROD_OAUTH_CLIENT_ID)
    PROD_CLIENT_SECRET: $(PROD_OAUTH_CLIENT_SECRET)
  inputs:
    script: |
      flashpipe pd-deploy --config flashpipe-pd-prod.yml --dry-run
```

## Configuration Precedence

Settings are applied in this order (later overrides earlier):

1. Config file defaults
2. Config file settings under `pd-snapshot:` or `pd-deploy:`
3. Command-line flags

**Example:**
```yaml
# flashpipe-pd.yml
pd-deploy:
  resources-path: ./cpars
  replace: true
  full-sync: true
```

```bash
# Command-line flag overrides config file
flashpipe pd-deploy --config flashpipe-pd.yml --full-sync=false
# Result: full-sync is false (from command line)
```

## Quick Reference

### Config File Keys

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `pd-snapshot.resources-path` | string | `./partner-directory` | Where to save files |
| `pd-snapshot.replace` | bool | `true` | Replace existing files |
| `pd-snapshot.pids` | list | `[]` | Filter specific PIDs |
| `pd-deploy.resources-path` | string | `./partner-directory` | Where to read files |
| `pd-deploy.replace` | bool | `true` | Replace existing values |
| `pd-deploy.full-sync` | bool | `false` | Delete remote params not in local |
| `pd-deploy.dry-run` | bool | `false` | Preview changes only |
| `pd-deploy.pids` | list | `[]` | Filter specific PIDs |

### Common Commands

```bash
# Snapshot with config
flashpipe pd-snapshot --config flashpipe-pd.yml

# Deploy with config
flashpipe pd-deploy --config flashpipe-pd.yml

# Deploy dry-run
flashpipe pd-deploy --config flashpipe-pd.yml --dry-run

# Override resources path
flashpipe pd-deploy --config flashpipe-pd.yml --resources-path ./other-dir

# Deploy specific PIDs only
flashpipe pd-deploy --config flashpipe-pd.yml --pids "PID1,PID2"
```

## Best Practices

1. **Use Environment Variables for Secrets**
   ```yaml
   oauth-clientid: ${OAUTH_CLIENT_ID}
   oauth-clientsecret: ${OAUTH_CLIENT_SECRET}
   ```

2. **One Config Per Environment**
   - `flashpipe-pd-dev.yml`
   - `flashpipe-pd-qa.yml`
   - `flashpipe-pd-prod.yml`

3. **Enable Full Sync Only in Upper Environments**
   ```yaml
   # Dev: full-sync: false (safer)
   # QA: full-sync: true (can match prod)
   # Prod: full-sync: true (authoritative)
   ```

4. **Always Dry-Run Production First**
   ```bash
   flashpipe pd-deploy --config prod.yml --dry-run
   # Review output
   flashpipe pd-deploy --config prod.yml
   ```

5. **Version Control Your Configs**
   - Store in Git alongside partner directory files
   - Review changes in pull requests
   - Use GitOps workflow

6. **Use PID Filters for Sensitive Partners**
   ```yaml
   pd-deploy:
     pids:
       - CRITICAL_PARTNER_001
       - SENSITIVE_SYSTEM_002
   ```
