# Documentation
The key components of _FlashPipe_ are
- **Java executables** - provide native access to SAP Integration Suite APIs
- **Unix scripts** - provide simplified access to SAP Integration Suite APIs
- **Local Maven repository** - provides cached libraries for Maven-based unit testing in offline mode

_FlashPipe_ uses the [public APIs of the SAP Integration Suite](https://api.sap.com/package/CloudIntegrationAPI?section=Artifacts) to automate the Build-To-Deploy cycle. The components are implemented in Groovy and compiled as Java executables.

While it is possible to use the Java executables directly, the Unix scripts do most of the heavy lifting by orchestrating between the various API calls required to complete the Build-To-Deploy cycle.

## Prerequisite
To use _FlashPipe_, you will need the following
1. Access to **Cloud Integration** on an SAP Integration Suite tenant - typically an Integration Developer credentials are required
2. Access to a **CI/CD platform**, e.g. Azure Pipelines, GitHub Actions
3. **Git-based repository** to host the contents of the Cloud Integration artifacts

Technically, it should be possible to use _FlashPipe_ on any CI/CD platform that supports container-based pipeline execution and Unix script execution.

## Usage of Unix scripts
For details on usage of the Unix scripts in pipeline steps, visit the [Unix scripts page](unix-scripts.md).

## Setup examples
Following are examples of setting up _FlashPipe_ in different CI/CD platforms.
- [Azure Pipelines](azure-pipelines.md)