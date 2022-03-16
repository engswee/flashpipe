# Release Notes

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

- 🔥🔥 New feature 🔥🔥 - _FlashPipe_ now supports [testing using simulation mode](simulation-testing.md) in Neo environments. While this is just a one-liner in this release notes, it is a power-packed feature! 😎

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

- 🔥🔥 New feature 🔥🔥 - _FlashPipe_ now enables [syncing of integration flow contents](github-actions-sync.md) from the tenant to Git repository
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