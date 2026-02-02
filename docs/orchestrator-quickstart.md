# Orchestrator Quick Start Guide

Get started with the Flashpipe Orchestrator in 5 minutes.

## Prerequisites

- Flashpipe installed
- SAP CPI tenant credentials
- Integration packages in local directory

## Step 1: Create Config File

Create `$HOME/flashpipe.yaml` with your tenant credentials:

```yaml
tmn-host: your-tenant.hana.ondemand.com
oauth-host: your-tenant.authentication.sap.hana.ondemand.com
oauth-clientid: your-client-id
oauth-clientsecret: your-client-secret
```

**Windows:** `%USERPROFILE%\flashpipe.yaml`
**Linux/macOS:** `$HOME/flashpipe.yaml`

## Step 2: Create Deployment Config

Create `001-deploy-config.yml` in your project root:

```yaml
deploymentPrefix: "DEV"  # Optional: adds prefix to all packages/artifacts

packages:
  - integrationSuiteId: "MyPackage"
    packageDir: "MyPackage"
    displayName: "My Integration Package"
    sync: true
    deploy: true
    
    artifacts:
      - artifactId: "MyIntegrationFlow"
        artifactDir: "MyIntegrationFlow"
        displayName: "My Integration Flow"
        type: "IntegrationFlow"
        sync: true
        deploy: true
```

## Step 3: Organize Your Packages

Ensure your directory structure looks like this:

```
.
├── packages/
│   └── MyPackage/              # Matches packageDir above
│       └── MyIntegrationFlow/  # Matches artifactDir above
│           ├── META-INF/
│           │   └── MANIFEST.MF
│           └── src/
│               └── main/
│                   └── resources/
│                       └── parameters.prop
└── 001-deploy-config.yml
```

## Step 4: Run the Orchestrator

```bash
# Update and deploy
flashpipe orchestrator --update --deploy-config ./001-deploy-config.yml

# Or update only (no deployment)
flashpipe orchestrator --update-only --deploy-config ./001-deploy-config.yml

# Or deploy only (no updates)
flashpipe orchestrator --deploy-only --deploy-config ./001-deploy-config.yml
```

## Common Use Cases

### Deploy to Different Environments

```bash
# Deploy to DEV
flashpipe orchestrator --update \
  --deployment-prefix DEV \
  --deploy-config ./deploy-config.yml

# Deploy to QA
flashpipe orchestrator --update \
  --deployment-prefix QA \
  --deploy-config ./deploy-config.yml

# Deploy to PROD
flashpipe orchestrator --update \
  --deployment-prefix PROD \
  --deploy-config ./deploy-config.yml
```

### Deploy Only Specific Packages

```bash
flashpipe orchestrator --update \
  --package-filter "MyPackage,OtherPackage" \
  --deploy-config ./deploy-config.yml
```

### Deploy Only Specific Artifacts

```bash
flashpipe orchestrator --update \
  --artifact-filter "MyIntegrationFlow,CriticalFlow" \
  --deploy-config ./deploy-config.yml
```

### Override Parameters

Add `configOverrides` to your artifact configuration:

```yaml
artifacts:
  - artifactId: "MyIntegrationFlow"
    artifactDir: "MyIntegrationFlow"
    type: "IntegrationFlow"
    configOverrides:
      SenderURL: "https://qa.example.com/api"
      Timeout: "60000"
      RetryCount: "3"
```

### Enable Debug Logging

```bash
flashpipe orchestrator --update --deploy-config ./deploy-config.yml --debug
```

## Without Config File

If you prefer not to use a config file:

```bash
flashpipe orchestrator --update \
  --deploy-config ./deploy-config.yml \
  --tmn-host your-tenant.hana.ondemand.com \
  --oauth-host your-tenant.authentication.sap.hana.ondemand.com \
  --oauth-clientid your-client-id \
  --oauth-clientsecret your-client-secret
```

## Troubleshooting

### "Required flag not set"

Make sure you have either:
- A config file at `$HOME/flashpipe.yaml`, OR
- Use `--config /path/to/config.yaml`, OR
- Provide all connection flags (`--tmn-host`, `--oauth-host`, etc.)

### "Package directory not found"

Check that:
- `packageDir` in your config matches the actual folder name
- You're running the command from the correct directory
- The path in `--packages-dir` is correct (default: `./packages`)

### "Artifact update failed"

Enable debug mode to see detailed logs:
```bash
flashpipe orchestrator --update --debug --deploy-config ./deploy-config.yml
```

### "Deployment failed"

- Check that artifacts updated successfully first
- Verify artifact has no validation errors in CPI
- Check CPI tenant logs for detailed error messages
- Try deploying individual artifacts to isolate the issue

## Next Steps

- See [full documentation](./orchestrator.md) for all features
- Learn about [multi-source configs](./orchestrator.md#configuration-sources)
- Set up [CI/CD integration](./orchestrator.md#cicd-integration)
- Generate configs automatically with [`config-generate`](./config-generate.md)

## Example: Complete Workflow

```bash
# 1. Generate deployment config from existing packages
flashpipe config-generate --packages-dir ./packages --output ./deploy-config.yml

# 2. Review and customize the generated config
nano deploy-config.yml

# 3. Deploy to DEV environment
flashpipe orchestrator --update \
  --deployment-prefix DEV \
  --deploy-config ./deploy-config.yml

# 4. If successful, deploy to QA
flashpipe orchestrator --update \
  --deployment-prefix QA \
  --deploy-config ./deploy-config.yml

# 5. Finally, deploy to PROD
flashpipe orchestrator --update \
  --deployment-prefix PROD \
  --deploy-config ./deploy-config.yml
```

## Tips

1. **Use version control** for your deployment configs
2. **Test in DEV first** with a deployment prefix
3. **Use filters** during development to deploy only what you're working on
4. **Keep credentials secure** - use config files instead of command-line flags
5. **Enable debug mode** when troubleshooting issues
6. **Use `--update-only`** first to verify changes before deploying
7. **Leverage config generation** to bootstrap new projects

## Need Help?

- Full documentation: [orchestrator.md](./orchestrator.md)
- Migration guide: [ORCHESTRATOR_MIGRATION.md](../ORCHESTRATOR_MIGRATION.md)
- Partner Directory: [partner-directory.md](./partner-directory.md)
- GitHub Issues: [Report a bug or request a feature](https://github.com/engswee/flashpipe/issues)