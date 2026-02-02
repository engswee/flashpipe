<img src="https://github.com/engswee/flashpipe/raw/main/docs/images/logo/flashpipe_logo_wording.png" alt="FlashPipe Logo" width="200" height="140"/>

## The CI/CD Companion for SAP Integration Suite

[![Build and Deploy](https://github.com/engswee/flashpipe/actions/workflows/go-prod.yml/badge.svg)](https://github.com/engswee/flashpipe/actions/workflows/go-prod.yml)
[![GitHub license](https://img.shields.io/github/license/engswee/flashpipe)](https://github.com/engswee/flashpipe/blob/main/LICENSE)
[![GitHub release](https://img.shields.io/github/release/engswee/flashpipe.svg)](https://github.com/engswee/flashpipe/releases/latest)
[![Docker Image Size (latest semver)](https://img.shields.io/docker/image-size/engswee/flashpipe)](https://hub.docker.com/r/engswee/flashpipe/tags?page=1&ordering=last_updated)
[![Docker Pulls](https://img.shields.io/docker/pulls/engswee/flashpipe)](https://hub.docker.com/r/engswee/flashpipe/tags?page=1&ordering=last_updated)

### About

_FlashPipe_ is a public [Docker image](https://hub.docker.com/r/engswee/flashpipe) that provides Continuous
Integration (CI) & Continuous Delivery/Deployment (CD) capabilities for SAP Integration Suite.

_FlashPipe_ aims to simplify the Build-To-Deploy cycle for SAP Integration Suite by providing CI/CD capabilities for
automating time-consuming manual tasks.

### Enhanced Capabilities

_FlashPipe_ has been significantly enhanced with powerful new commands for streamlined CI/CD workflows:

#### 🎯 Orchestrator Command

High-level deployment orchestration with integrated workflow management:

- **Complete Lifecycle**: Update and deploy packages and artifacts in a single command
- **Multi-Source Configs**: Load from files, folders, or remote URLs
- **YAML Configuration**: Define all settings in a config file for reproducibility
- **Parallel Deployment**: Deploy multiple artifacts simultaneously (3-5x faster)
- **Environment Support**: Multi-tenant/environment prefixes (DEV, QA, PROD)
- **Selective Processing**: Filter by specific packages or artifacts

```bash
# Simple deployment with YAML config
flashpipe orchestrator --orchestrator-config ./orchestrator.yml

# Or with individual flags
flashpipe orchestrator --update \
  --deployment-prefix DEV \
  --deploy-config ./001-deploy-config.yml \
  --packages-dir ./packages
```

#### ⚙️ Config Generation

Automatically generate deployment configurations from your packages directory:

```bash
# Generate config from package structure
flashpipe config-generate --packages-dir ./packages --output ./deploy-config.yml
```

#### 📁 Partner Directory Management

Snapshot and deploy Partner Directory parameters:

```bash
# Download parameters from SAP CPI
flashpipe pd-snapshot --output ./partner-directory

# Upload parameters to SAP CPI
flashpipe pd-deploy --source ./partner-directory
```

See documentation below for complete details on each command.

### Documentation

For comprehensive documentation on using _FlashPipe_, visit the [GitHub Pages documentation site](https://engswee.github.io/flashpipe/).

#### New Commands Documentation

- **[Orchestrator](docs/orchestrator.md)** - High-level deployment orchestration and workflow management
- **[Orchestrator Quick Start](docs/orchestrator-quickstart.md)** - Get started with orchestrator in 30 seconds
- **[Orchestrator YAML Config](docs/orchestrator-yaml-config.md)** - Complete YAML configuration reference
- **[Configure](docs/configure.md)** - Configure artifact parameters with YAML files
- **[Config Generate](docs/config-generate.md)** - Automatically generate deployment configurations
- **[Partner Directory](docs/partner-directory.md)** - Manage Partner Directory parameters

#### Migration Guides

- **[Orchestrator Migration Guide](docs/orchestrator-migration.md)** - Migrate from standalone CLI to integrated orchestrator

#### Core FlashPipe Documentation

- **[FlashPipe CLI Reference](docs/flashpipe-cli.md)** - Complete CLI command reference
- **[OAuth Client Setup](docs/oauth_client.md)** - Configure OAuth authentication
- **[GitHub Actions Integration](docs/documentation.md)** - CI/CD pipeline examples

#### Examples

Configuration examples are available in [docs/examples/](docs/examples/):
- `orchestrator-config-example.yml` - Orchestrator configuration template
- `flashpipe-cpars-example.yml` - Partner Directory configuration example

#### Developer Documentation

For contributors and maintainers, see [dev-docs/](dev-docs/) for:
- Testing guides and coverage reports
- CLI porting summaries
- Enhancement documentation

### Analytics

_FlashPipe_ collects anonymous usage analytics to help guide ongoing development and improve the tool. No personal or user-identifiable information is collected. Analytics collection is always enabled and not optional. By using _FlashPipe_, you consent to this data collection. If you have concerns, you are encouraged to review the implementation. If you do not agree, please refrain from using the tool.

### Examples Repository
The following repository on GitHub provides examples of different use cases of _FlashPipe_.

[https://github.com/engswee/flashpipe-demo](https://github.com/engswee/flashpipe-demo)

### Versioning
[SemVer](https://semver.org/) is used for versioning.

### Contributing

Contributions from the community are welcome.

To contribute to _FlashPipe_, check the [contribution guidelines page](CONTRIBUTING.md).

### License

_FlashPipe_ is licensed under the terms of Apache License, Version 2.0 - see the [LICENSE](LICENSE) file for details.

### Stargazers over time
[![Stargazers over time](https://starchart.cc/engswee/flashpipe.svg?variant=adaptive)](https://starchart.cc/engswee/flashpipe)


