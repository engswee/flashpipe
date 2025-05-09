# Release Notes

## 3.5.1 (Released 9 May 2025)

---

- Bug fix - update handling of dynamic expressions found in externalised parameters
- Update to latest Go version 1.24.3 and dependencies to resolve following security vulnerabilities
  - CVE-2025-22869
  - CVE-2025-22872
  - CVE-2025-22870

## 3.5.0 (Released 20 December 2024)

---

- 🔥🔥 Hot new feature 🔥🔥 - _FlashPipe_ now supports restoring snapshot from Git repository to tenant
  - New `snapshot restore` command
    - Handles creating/updating multiple CPI integration packages (and corresponding artifacts) from snapshot in Git repository to tenant
    - Refer to [FlashPipe CLI page](flashpipe-cli.md) for more details
- Remove values for `--target` flag of `sync` and `sync apim` commands
  - Previously deprecated `local` & `remote` values are no longer allowed
- Update default value for `--sync-package-details` flag of `snapshot` command to `true`
- Update behavior of `sync` command when integration package does not exist in tenant
  - If there is a JSON file with the integration package details in the directory, then it is used to create the integration package first, without requiring execution of `update package` command
- Update to latest Go version 1.23.4 and dependencies to resolve following security vulnerabilities
  - CVE-2024-45337
  - CVE-2024-45338

## 3.4.1 (Released 19 December 2024)

---

- Bug fix - update handling when runtime artifact has not been deployed (due to change in behaviour of IntegrationRuntimeArtifacts API by SAP)

## 3.4.0 (Released 28 June 2024)

---

- New flags for `snapshot` command
  - New flag `--dir-artifacts` (env var `FLASHPIPE_DIR_ARTIFACTS`)
    - allows the artifact contents to be stored in a subdirectory of the Git repository (grouped into packaged)
  - New flags `--ids-include` (env var `FLASHPIPE_IDS_INCLUDE`) and `--ids-exclude` (env var `FLASHPIPE_IDS_EXCLUDE`)
    - allows including or excluding specific package IDs from snapshot
- Supporting Shell expansion for environment variables for certain parameters (details in [FlashPipe CLI page](flashpipe-cli.md))
  - contributed by [Vadim Klimov](https://github.com/vadimklimov)

## 3.3.0 (Released 17 May 2024)

---

- Include designtime artifact's description when performing update/sync between tenant and Git
- When updating designtime artifacts in tenant, the name always comes from MANIFEST.MF
- Include warning for errors when syncing APIM, commonly due to using CPI tenant and credentials incorrectly
- Introduce delay when there is error in deployment so that the error details are returned correctly
- Bug fix - correctly sync updates of integration package details from tenant to Git

## 3.2.3 (Released 1 May 2024)

---

- Bug fix - update handling of artifact name when running `update artifact` or `sync` to tenant, due to SAP changing width of MANIFEST.MF file to 72 characters
- Update to latest dependencies to resolve following security vulnerabilities
  - CVE-2024-24786
  - CVE-2023-45288

## 3.2.2 (Released 11 March 2024)

---

- Improve error handling when unmarshalling of JSON fails by displaying content in logs

## 3.2.1 (Released 31 January 2024)

---

- Bug fix - correct handling of artifact name when running `sync` command to tenant - when the name is longer than the width of MANIFEST.MF, then only a single space is trimmed from the next line

## 3.2.0 (Released 25 January 2024)

---

- 🔥🔥 Hot new feature 🔥🔥 - _FlashPipe_ now supports syncing of **API Management** artifacts 
  - New `sync apim` command
  - Handles syncing of API Proxies and dependent artifacts
- Incorporation of usage analytics - no sensitive/personal data collected
- Rename allowed values for `--target` flag of `sync` command so that it provides more clarity
  - `local` & `remote` are now deprecated
  - `git` & `tenant` are the new replacement values
- New feature for `update artifact` command to default artifact name to value from MANIFEST.MF when optional flag `--artifact-name` is not provided
- Update to latest dependencies to resolve following security vulnerabilities
  - CVE-2023-48795
  - CVE-2023-49569
  - CVE-2023-49568
  - GHSA-9763-4f94-gfch

## 3.1.0 (Released 9 October 2023)

---

- 🔥 New feature 🔥 - _FlashPipe_ now supports syncing from Git repository to tenant
  - `sync` command
    - New flag `--target` (env var `FLASHPIPE_TARGET`) enables sync of multiple artifacts from Git repository to tenant
- Changes for `update artifact` command
  - Flags `--artifact-name` & `--package-name` are now optional to simplify configuration. When not provided, they will default to the value of their corresponding IDs.

## 3.0.0 (Released 4 September 2023)

---

- 🔥🔥 New major release 🔥🔥 - _FlashPipe_ reimplemented fully in [Go](https://go.dev) 
  - Additional support for the following artifact types
    - Message Mapping
    - Script Collection
    - Value Mapping
  - A ✨✨ shiny new CLI ✨✨ built on [Cobra](https://cobra.dev), and together with [Viper](https://github.com/spf13/viper), configuration parameters can be passed with the following methods
    - CLI flags
    - Environment variables
    - Config file (flashpipe.yaml)
  - Consolidate to just a single Docker image, with the introduction of `latest` rolling tag
  - Reduced Docker image size of up to 60% compared to the previous lite image
  - Names of environment variables have changed, refer to [FlashPipe CLI page](flashpipe-cli.md)
  - `sync` and `snapshot` commands
    - New parameters for configuring username and email when committing changes to Git repository
  - Support for Neo is dropped due to sunset of the platform as described [here](https://blogs.sap.com/2023/06/14/farewell-neo-sap-btp-multi-cloud-environment-the-deployment-environment-of-choice/)

## 2.7.1 (Released 17 January 2023)

---

- Bug fix - handle issues with syncing to Git due to new behavior introduced in latest version of git in Debian image

## 2.7.0 (Released 9 January 2023)

---

- 🔥🔥 New feature 🔥🔥 - _FlashPipe_ now supports syncing and updating integration package details
  - New script `update_integration_package.sh` can be configured in a pipeline for updating integration package details on tenant
  - Existing script `sync_to_git_repository.sh` can be configured to sync integration package details from a tenant to a Git repository
    - New optional parameters `SYNC_PACKAGE_LEVEL_DETAILS`, `NORMALIZE_PACKAGE_ACTION`, `NORMALIZE_PACKAGE_ID_PREFIX_SUFFIX`, `NORMALIZE_PACKAGE_NAME_PREFIX_SUFFIX` to handle syncing of integration package
  - Existing script `snapshot_to_git_repository.sh` can be configured to sync integration package details from a tenant to a Git repository
    - New optional parameters `SYNC_PACKAGE_LEVEL_DETAILS` to handle syncing of integration package

## 2.6.1 (Released 31 October 2022)

---

- Minor fix to ignore Origin-* properties in MANIFEST.MF files when performing comparison during sync or update

## 2.6.0 (Released 31 March 2022)

---

- Deprecated parameters `MANIFEST_FILE` and `VERSION_HANDLING` for `update_designtime_artifact.sh`
  - `MANIFEST_FILE` - it will now always refer to the file `META-INF/MANIFEST.MF`, and automatically convert based on provided `IFLOW_ID` and `IFLOW_NAME` 
  - `VERSION_HANDLING` - it will now behave as though `VERSION_HANDLING` = `MANIFEST`, meaning version number will depend on `Bundle-Version` set in `META-INF/MANIFEST.MF` file
- New optional parameter for `deploy_runtime_artifact.sh` 
  - `COMPARE_VERSIONS` - perform version comparison between design time artifact and runtime before deployment (Default = `true`)
- New optional parameters for `sync_to_git_repository.sh`
  - `SCRIPT_COLLECTION_MAP` - handle conversion of script collection references in `MANIFEST.MF` and IFlow BPMN XML files
  - `NORMALIZE_MANIFEST_ACTION` - normalize IFlow ID and IFlow Name in `MANIFEST.MF` by adding or deleting prefix/suffix
  - `NORMALIZE_MANIFEST_PREFIX_SUFFIX` - prefix or suffix used in normalization of IFlow ID and IFlow Name in `MANIFEST.MF`

## 2.5.3 (Released 16 March 2022)

---

- Further bug fix in 2.5.2 when `SCRIPT_COLLECTION_MAP` is empty during IFlow update

## 2.5.2 (Released 16 March 2022)

---

- Fix bug in 2.5.1 when `SCRIPT_COLLECTION_MAP` is empty during IFlow update

## 2.5.1 (Released 15 March 2022)

---

- Include META-INF/MANIFEST.MF file for diff comparison during sync to Git, and update to tenant

## 2.5.0 (Released 28 February 2022)

---

- New functionalities during upload or update of Integration Flows
  - Automatic update of following attributes in MANIFEST.MF from environment variables
    - `Bundle-SymbolicName` - from `IFLOW_ID`
    - `Bundle-Name` - from `IFLOW_NAME`
  - New optional parameter `SCRIPT_COLLECTION_MAP` to handle conversion of script collection references in MANIFEST.MF and IFlow BPMN XML files

## 2.4.7 (Released 5 January 2022)

---

- Bug fix - bump log4j version to 2.17.1 due to security vulnerability as described in [CVE-2021-44832](https://nvd.nist.gov/vuln/detail/CVE-2021-44832)

## 2.4.6 (Released 20 December 2021)

---

- Bug fix - bump log4j version to 2.17.0 due to security vulnerability as described in [CVE-2021-45105](https://nvd.nist.gov/vuln/detail/CVE-2021-45105)

## 2.4.5 (Released 16 December 2021)

---

- Bug fix - bump log4j version to 2.16.0 due to security vulnerability as described in [CVE-2021-45046](https://nvd.nist.gov/vuln/detail/CVE-2021-45046)

## 2.4.4 (Released 13 December 2021)

---

- Bug fix - bump log4j version to 2.15.0 due to security vulnerability as described in [CVE-2021-44228](https://nvd.nist.gov/vuln/detail/CVE-2021-44228)

## 2.4.3 (Released 20 October 2021)

---

- Stability release - no new features, but added lots of unit tests and integration tests to ensure good code coverage during testing

## 2.4.2 (Released 21 September 2021)

---

- Bug fix - add delay before checking deployment status of IFlow to allow time for the deployment to kick in

## 2.4.1 (Released 7 September 2021)

---

- `deploy_runtime_artifact.sh` now supports deployment of multiple IFlows in a single step
  - Parameter `IFLOW_ID` accepts a comma separated list of IFlow IDs

## 2.4.0 (Released 18 August 2021)

---

- 🔥🔥 New feature 🔥🔥 - _FlashPipe_ now supports testing using simulation mode in Neo environments. While this is just a one-liner in this release notes, it is a power-packed feature! 😎

## 2.3.0 (Released 2 August 2021)

---

- 🔥🔥 New feature 🔥🔥 - _FlashPipe_ now enables [snapshot of tenant's integration packages & flows](github-actions-snapshot.md) to Git repository - courtesy of contribution from [Ariel Bravo Ayala](https://github.com/ambravo)
  - New script `snapshot_to_git_repository.sh` can be configured in a pipeline for periodic or adhoc sync
- New optional parameters for `update_designtime_artifact.sh`
  - `VERSION_HANDLING` - handling of version number when updating IFlow (automatic increment patch no, or based on MANIFEST.MF file)
- Introducing new rolling tags for Docker images (for non-production usage) - another brilliant suggestion from [Ariel Bravo Ayala](https://github.com/ambravo)
  - `2.x.x` & `2.x.x-lite` - rolling tag that always refer to the latest release of major version 2
  - `2.3.x` & `2.3.x-lite` - rolling tag that always refer to the latest release of minor version 2.3

## 2.2.1 (Released 9 July 2021)

---

- Fix bug related to recursive directory creation during syncing of IFlows 
- Add validation to certain input environment variables to check that they do not contain secrets

## 2.2.0 (Released 7 July 2021)

---

- Added new optional environment variable `HOST_OAUTH_PATH` to handle OAuth authentication with Neo environments 
- [OAuth client setup page](oauth_client.md) updated to include steps for Neo environment

## 2.1.1 (Released 1 July 2021)

---

- Switch environment variable `GIT_DIR` to `GIT_SRC_DIR` for `update_designtime_artifact.sh` due to conflict with default Git variable

## 2.1.0 (Released 30 June 2021)

---

- 🔥🔥 New feature 🔥🔥 - _FlashPipe_ now enables [syncing of integration flow contents](github-actions-sync-to-git.md) from the tenant to Git repository
  - New script `sync_to_git_repository.sh` can be configured in a pipeline for periodic or adhoc sync
- Clean up the logs generated
  - A new default pattern layout for a cleaner simpler look
  - <span style="color:blue">Spice</span> <span style="color:green">up</span> log <span style="color:red">levels</span> <span style="color:orange">with</span> <span style="color:purple">color</span>
  - Add emojis 🛑 🏆 ⚠️ 🚀 to highlight key log messages

## 2.0.1 (Released 14 June 2021)

---

- Corrected handling of configuration parameters update via Configurations API

## 2.0.0 (Released 10 June 2021)

---

What, version 2.0.0 already?! Yes, there are incompatible changes related to passing input values to the Unix scripts, so according to [SemVer](https://semver.org), this bumps the MAJOR version up.
- Passing input values to the Unix script have been switched from command line arguments to environment variables
- Additional support for authentication using OAuth when accessing the APIs
- There is now a corresponding lite Docker image for each version, `<version_no>-lite`. This image is smaller in size and does not contain the full Maven capabilities

## 1.0.2 (Released 25 May 2021)

---

- Bug fix for error in comparison before deployment if there are no runtime artifact

## 1.0.1 (Released 22 May 2021)

---

- New configurable parameters for `deploy_runtime_artifact.sh`
    - delay - delay between each check of IFlow deployment status
    - maxcheck - max limit for number of times to check IFlow deployment status
- Compare designtime version with runtime version before deployment

## 1.0.0 (Released 20 May 2021)

---
Initial Release of _FlashPipe_ with the following features:

- Create/Update designtime artifacts (Integration Flow) on Cloud Integration
- Deploy designtime artifacts to Cloud Integration runtime
- Automatic creation of Integration Package if it does not exist
- Automatic comparison of artifact contents from Git repository against tenant
- Artifact creation, update, deployment across multiple environments
    - Multiple copies on same tenant (with different IDs and configuration values)
    - Deployment on different tenants (Dev/QA vs Prod)