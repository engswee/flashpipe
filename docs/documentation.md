# Documentation
The key components of _FlashPipe_ are
- **Go binary executable** - a Unix-based CLI which provides commands to interact with an SAP Integration Suite tenant
- **Docker image** - fully contained image that can be used to execute _FlashPipe_ commands in a CI/CD pipeline
- **Companion GitHub Action** [engswee/flashpipe-action](https://github.com/engswee/flashpipe-action) - custom action to simplify usage in GitHub Actions workflows

_FlashPipe_ uses the public APIs of the SAP Integration Suite ([Cloud Integration](https://api.sap.com/package/CloudIntegrationAPI/odata) & [API Management](https://api.sap.com/package/APIMgmt/odata)) to automate the Build-To-Deploy cycle. The components are implemented in Go and compiled as Unix executables.

## Prerequisite
To use _FlashPipe_, you will need the following
1. Access to **Cloud Integration** on an SAP Integration Suite tenant - typically an Integration Developer permissions are required
2. Access to **API Management** on an SAP Integration Suite tenant - API access only using OAuth
3. Access to a **CI/CD platform**, e.g. [Azure Pipelines](https://azure.microsoft.com/en-us/services/devops/pipelines/), [GitHub Actions](https://github.com/features/actions)
4. **Git-based repository** to host the contents of the Cloud Integration artifacts

Technically, it should be possible to use _FlashPipe_ on any CI/CD platform that supports container-based pipeline execution.

## Docker image tags
With major release `3.0.0`, _FlashPipe_ returns to the simpler approach of having a single tag for each version release. Separate full and lite tags are no longer available.
The latest Docker image for _FlashPipe_ is

  `engswee/flashpipe:3.2.0`

For a list of all available images tags, check [here](https://hub.docker.com/r/engswee/flashpipe/tags).

### Rolling tags
Starting from version `3.0.0`, rolling tag `latest` is introduced to make it easier to get the latest version. This rolling tag is dynamic and will point to the latest version of the image.

### Usage recommendation
- When using _FlashPipe_ in productive pipelines, use an immutable tag (e.g. `3.2.0`) to ensure stability so that the pipeline will not be affected negatively by new version releases.
- When using _FlashPipe_ in development pipelines, use rolling tag `latest` to always get the latest version.

## Authentication
_FlashPipe_ supports the following methods of authentication when accessing the SAP Integration Suite APIs.
- Basic authentication (only for Cloud Integration)
- OAuth authentication

It is recommended to use OAuth so that the access is not linked to an individual's credential (which may be revoked or the password might change). For details on setting up an OAuth client for use with _FlashPipe_, visit the [OAuth client setup page](oauth_client.md).

## Usage of CLI
For details on usage of the CLI commands in pipeline steps, visit the [Flashpipe CLI page](flashpipe-cli.md).

## Usage examples
Following are different usage examples of _FlashPipe_ on different CI/CD platforms.
- [Upload/Deploy designtime artifacts using Azure Pipelines](azure-pipelines-upload.md)
- [Upload/Deploy designtime artifacts using GitHub Actions](github-actions-upload.md)
- [Sync designtime artifacts from Tenant to GitHub using GitHub Actions](github-actions-sync-to-git.md)
- [Sync designtime artifacts from GitHub to Tenant using GitHub Actions](github-actions-sync-to-tenant.md)
- [Snapshot Tenant Content to GitHub using GitHub Actions](github-actions-snapshot.md)
- [Sync APIM artifacts between Tenant and GitHub using GitHub Actions](github-actions-sync-apim.md)

## Archive
For older versions of _FlashPipe_ that are implemented in Java/Groovy, refer to the [archive](https://github.com/engswee/flashpipe/tree/archive) branch of this repository.

## Reference
The following repository on GitHub provides sample usage of _FlashPipe_.

[https://github.com/engswee/flashpipe-demo](https://github.com/engswee/flashpipe-demo)