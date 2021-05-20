# Release Notes

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