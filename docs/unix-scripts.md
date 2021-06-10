# Unix Scripts in FlashPipe

_FlashPipe_ provides the following Unix scripts for accessing SAP Integration Suite APIs.
- **update_designtime_artifact.sh**
- **deploy_runtime_artifact.sh**

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
    GIT_DIR - directory containing contents of Integration Flow

Optional environment variables:
    PARAM_FILE - Use to a different parameters.prop file instead of the default in src/main/resources/
    MANIFEST_FILE - Use to a different MANIFEST.MF file instead of the default in META-INF/
    WORK_DIR - Working directory for in-transit files (default is /tmp if not set)

NOTE: Encapsulate values in double quotes ("") if there are space characters in them
```

#### Example
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
    GIT_DIR: "FlashPipe Demo/Groovy XML Transformation"
```

### 2. deploy_runtime_artifact.sh
This script is used to deploy a Cloud Integration designtime artifact to the runtime. It will compare the version of the designtime artifact against the runtime artifact before executing deployment if there are diferences.


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
    IFLOW_ID - ID of Integration Flow

Optional environment variables:
    DELAY_LENGTH - Delay (in seconds) between each check of IFlow deployment status (default to 30 if not set)
    MAX_CHECK_LIMIT - Max number of times to check for IFlow deployment status (default to 10 if not set)
```

#### Example
```bash
/usr/bin/deploy_runtime_artifact.sh

Environment variables set before call:
    HOST_TMN: ***.hana.ondemand.com
    HOST_OAUTH: ***.authentication.<region>.hana.ondemand.com
    OAUTH_CLIENTID: <clientid>
    OAUTH_CLIENTSECRET: <clientsecret>
    IFLOW_ID: GroovyXMLTransformation
```