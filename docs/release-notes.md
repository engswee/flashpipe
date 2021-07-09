# Release Notes

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