# Unix Scripts in FlashPipe

_FlashPipe_ provides the following Unix scripts for accessing SAP Integration Suite APIs.
- **update_designtime_artifact.sh**
- **deploy_runtime_artifact.sh**
- **sync_to_git_repository.sh**
- **snapshot_to_git_repository.sh**

These scripts perform the _magic_ that significantly simplifies the steps required to complete the build and deploy steps in a CI/CD pipeline.

The following section describes the usage of the scripts. Since version 2.0.0, input values are passed via environment variables instead of command line arguments.

### 1. update_designtime_artifact.sh
This script is used to create/update a Cloud Integration designtime artifact to the tenant. It provides the following functionalities:
- check existence of artifact to determine if it needs to be created or updated
- create Integration Package (if it does not exist) to store the artifact
- compare contents of artifact in Git repository against tenant to determine if artifact in tenant needs to be updated
- use different `MANIFEST.MF` and/or `parameters.prop` files to deploy different versions of the same artifact to same/different tenants
- create/update designtime artifact


#### Usage and environment variable list
```bash
/usr/bin/update_designtime_artifact.sh

Mandatory environment variables:
    HOST_TMN - Base URL for tenant management node of Cloud Integration (excluding the https:// prefix)
    BASIC_USERID - User ID (required when using Basic Authentication)
    BASIC_PASSWORD - Password (required when using Basic Authentication)
    HOST_OAUTH - Host name for OAuth authentication server (required when using OAuth Authentication, excluding the https:// prefix)
    OAUTH_CLIENTID - OAuth Client ID (required when using OAuth Authentication)
    OAUTH_CLIENTSECRET - OAuth Client Secret (required when using OAuth Authentication)
    IFLOW_ID - ID of Integration Flow
    IFLOW_NAME - Name of Integration Flow
    PACKAGE_ID - ID of Integration Package
    PACKAGE_NAME - Name of Integration Package
    GIT_SRC_DIR - Directory containing contents of Integration Flow

Optional environment variables:
    PARAM_FILE - Use to a different parameters.prop file instead of the default in src/main/resources/
    MANIFEST_FILE - [DEPRECATED] Use to a different MANIFEST.MF file instead of the default in META-INF/
    WORK_DIR - Working directory for in-transit files (default is /tmp if not set)
    HOST_OAUTH_PATH - Specific path for OAuth token server, e.g. example /oauth2/api/v1/token for Neo environments (default is /oauth/token if not set for CF environments)
    VERSION_HANDLING - [DEPRECATED] Determination of version number during artifact update
    SCRIPT_COLLECTION_MAP - Comma-separated source-target ID pairs for converting script collection references during upload/update

NOTE: Encapsulate values in double quotes ("") if there are space characters in them
```

#### Example (OAuth for Cloud Foundry)
```bash
/usr/bin/update_designtime_artifact.sh

Environment variables set before call:
    HOST_TMN: ***.hana.ondemand.com
    HOST_OAUTH: ***.authentication.<region>.hana.ondemand.com
    OAUTH_CLIENTID: <clientid>
    OAUTH_CLIENTSECRET: <clientsecret>
    IFLOW_ID: GroovyXMLTransformation
    IFLOW_NAME: "Groovy XML Transformation"
    PACKAGE_ID: FlashPipeDemo
    PACKAGE_NAME: "FlashPipe Demo"
    GIT_SRC_DIR: "FlashPipe Demo/Groovy XML Transformation"
    SCRIPT_COLLECTION_MAP: "DEV_Common_Scripts=Common_Scripts"
```

#### Example (OAuth for Neo)
```bash
/usr/bin/update_designtime_artifact.sh

Environment variables set before call:
    HOST_TMN: ***.hana.ondemand.com
    HOST_OAUTH: oauthasservices-<tenantid>.<region>.hana.ondemand.com
    HOST_OAUTH_PATH: /oauth2/api/v1/token
    OAUTH_CLIENTID: <clientid>
    OAUTH_CLIENTSECRET: <clientsecret>
    IFLOW_ID: GroovyXMLTransformation
    IFLOW_NAME: "Groovy XML Transformation"
    PACKAGE_ID: FlashPipeDemo
    PACKAGE_NAME: "FlashPipe Demo"
    GIT_SRC_DIR: "FlashPipe Demo/Groovy XML Transformation"
```

### 2. deploy_runtime_artifact.sh
This script is used to deploy Cloud Integration designtime artifact(s) to the runtime. It will compare the version of the designtime artifact against the runtime artifact before executing deployment if there are differences.


#### Usage and environment variable list
```bash
/usr/bin/deploy_runtime_artifact.sh

Mandatory environment variables:
    HOST_TMN - Base URL for tenant management node of Cloud Integration (excluding the https:// prefix)
    BASIC_USERID - User ID (required when using Basic Authentication)
    BASIC_PASSWORD - Password (required when using Basic Authentication)
    HOST_OAUTH - Host name for OAuth authentication server (required when using OAuth Authentication, excluding the https:// prefix)
    OAUTH_CLIENTID - OAuth Client ID (required when using OAuth Authentication)
    OAUTH_CLIENTSECRET - OAuth Client Secret (required when using OAuth Authentication)
    IFLOW_ID - Comma separated list of Integration Flow IDs

Optional environment variables:
    DELAY_LENGTH - Delay (in seconds) between each check of IFlow deployment status (default to 30 if not set)
    MAX_CHECK_LIMIT - Max number of times to check for IFlow deployment status (default to 10 if not set)
    HOST_OAUTH_PATH - Specific path for OAuth token server, e.g. example /oauth2/api/v1/token for Neo environments (default is /oauth/token if not set for CF environments)
```

#### Example (OAuth for Cloud Foundry)
```bash
/usr/bin/deploy_runtime_artifact.sh

Environment variables set before call:
    HOST_TMN: ***.hana.ondemand.com
    HOST_OAUTH: ***.authentication.<region>.hana.ondemand.com
    OAUTH_CLIENTID: <clientid>
    OAUTH_CLIENTSECRET: <clientsecret>
    IFLOW_ID: GroovyXMLTransformation
```

#### Example (OAuth for Neo)
```bash
/usr/bin/deploy_runtime_artifact.sh

Environment variables set before call:
    HOST_TMN: ***.hana.ondemand.com
    HOST_OAUTH: oauthasservices-<tenantid>.<region>.hana.ondemand.com
    HOST_OAUTH_PATH: /oauth2/api/v1/token
    OAUTH_CLIENTID: <clientid>
    OAUTH_CLIENTSECRET: <clientsecret>
    IFLOW_ID: GroovyXMLTransformation
```

### 3. sync_to_git_repository.sh
This script is used to sync Cloud Integration designtime artifacts from a tenant to a Git repository. It will compare any differences (new, deleted, changed) in files from tenant and commit/push to the Git repository.


#### Usage and environment variable list
```bash
/usr/bin/sync_to_git_repository.sh

Mandatory environment variables:
    HOST_TMN - Base URL for tenant management node of Cloud Integration (excluding the https:// prefix)
    BASIC_USERID - User ID (required when using Basic Authentication)
    BASIC_PASSWORD - Password (required when using Basic Authentication)
    HOST_OAUTH - Host name for OAuth authentication server (required when using OAuth Authentication, excluding the https:// prefix)
    OAUTH_CLIENTID - OAuth Client ID (required when using OAuth Authentication)
    OAUTH_CLIENTSECRET - OAuth Client Secret (required when using OAuth Authentication)
    PACKAGE_ID - ID of Integration Package
    GIT_SRC_DIR - Base directory containing contents of Integration Flow(s)

Optional environment variables:
    INCLUDE_IDS - List of included IFlow IDs
    EXCLUDE_IDS - List of excluded IFlow IDs
    DRAFT_HANDLING - Handling when IFlow is in draft version. Allowed values: SKIP (default), ADD, ERROR
    DIR_NAMING_TYPE - Name IFlow directories by ID or Name. Allowed values: ID (default), NAME
    COMMIT_MESSAGE - Message used in commit
    WORK_DIR - Working directory for in-transit files (default is /tmp if not set)
    HOST_OAUTH_PATH - Specific path for OAuth token server, e.g. example /oauth2/api/v1/token for Neo environments (default is /oauth/token if not set for CF environments)
```

#### Example (OAuth for Cloud Foundry)
```bash
/usr/bin/sync_to_git_repository.sh

Environment variables set before call:
    HOST_TMN: ***.hana.ondemand.com
    HOST_OAUTH: ***.authentication.<region>.hana.ondemand.com
    OAUTH_CLIENTID: <clientid>
    OAUTH_CLIENTSECRET: <clientsecret>
    PACKAGE_ID: FlashPipeDemo
    GIT_SRC_DIR: "FlashPipe Demo"
```

#### Example (OAuth for Neo)
```bash
/usr/bin/sync_to_git_repository.sh

Environment variables set before call:
    HOST_TMN: ***.hana.ondemand.com
    HOST_OAUTH: oauthasservices-<tenantid>.<region>.hana.ondemand.com
    HOST_OAUTH_PATH: /oauth2/api/v1/token
    OAUTH_CLIENTID: <clientid>
    OAUTH_CLIENTSECRET: <clientsecret>
    PACKAGE_ID: FlashPipeDemo
    GIT_SRC_DIR: "FlashPipe Demo"
```

### 4. snapshot_to_git_repository.sh
This script is used to capture a snapshot of the Cloud Integration tenant's artifacts to a Git repository. It will compare any differences (new, deleted, changed) in files from tenant and commit/push to the Git repository.


#### Usage and environment variable list
```bash
/usr/bin/snapshot_to_git_repository.sh

Mandatory environment variables:
    HOST_TMN - Base URL for tenant management node of Cloud Integration (excluding the https:// prefix)
    BASIC_USERID - User ID (required when using Basic Authentication)
    BASIC_PASSWORD - Password (required when using Basic Authentication)
    HOST_OAUTH - Host name for OAuth authentication server (required when using OAuth Authentication, excluding the https:// prefix)
    OAUTH_CLIENTID - OAuth Client ID (required when using OAuth Authentication)
    OAUTH_CLIENTSECRET - OAuth Client Secret (required when using OAuth Authentication)
    GIT_SRC_DIR - Base directory containing contents of artifacts (grouped into packages)

Optional environment variables:
    DRAFT_HANDLING - Handling when IFlow is in draft version. Allowed values: SKIP (default), ADD, ERROR
    COMMIT_MESSAGE - Message used in commit
    WORK_DIR - Working directory for in-transit files (default is /tmp if not set)
    HOST_OAUTH_PATH - Specific path for OAuth token server, e.g. example /oauth2/api/v1/token for Neo environments (default is /oauth/token if not set for CF environments)
```

#### Example (OAuth for Cloud Foundry)
```bash
/usr/bin/snapshot_to_git_repository.sh

Environment variables set before call:
    HOST_TMN: ***.hana.ondemand.com
    HOST_OAUTH: ***.authentication.<region>.hana.ondemand.com
    OAUTH_CLIENTID: <clientid>
    OAUTH_CLIENTSECRET: <clientsecret>
    GIT_SRC_DIR: "TrialTenant"
```

#### Example (OAuth for Neo)
```bash
/usr/bin/snapshot_to_git_repository.sh

Environment variables set before call:
    HOST_TMN: ***.hana.ondemand.com
    HOST_OAUTH: oauthasservices-<tenantid>.<region>.hana.ondemand.com
    HOST_OAUTH_PATH: /oauth2/api/v1/token
    OAUTH_CLIENTID: <clientid>
    OAUTH_CLIENTSECRET: <clientsecret>
    GIT_SRC_DIR: "TrialTenant"
```
