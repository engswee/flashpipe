# Unix Scripts in FlashPipe

_FlashPipe_ provides the following Unix scripts for accessing SAP Integration Suite APIs.
- **update_designtime_artifact.sh**
- **deploy_runtime_artifact.sh**

These scripts perform the _magic_ that significantly simplifies the steps required to complete the build and deploy steps in a CI/CD pipeline.

The following section describes the usage of the scripts.

### 1. update_designtime_artifact.sh
This script is used to create/update a Cloud Integration designtime artifact to the tenant. It provides the following functionalities:
- check existence of artifact to determine if it needs to be created or updated
- create Integration Package (if it does not exist) to store the artifact
- compare contents of artifact in Git repository against tenant to determine if artifact in tenant needs to be updated
- use different `MANIFEST.MF` and/or `parameters.prop` files to deploy different versions of the same artifact to same/different tenants
- create/update designtime artifact


#### Usage and parameters list
```bash
/usr/bin/update_designtime_artifact.sh [--param_file=<path_to_file>] [--manifest_file=<path_to_file>] [--logcfgfile=<path_to_file>] [--debug] <working_dir> <tmn_host> <cpi_user> <cpi_password> <artifact_id> <artifact_name> <package_id> <package_name> <git_src_dir>

Mandatory parameters (ensure order of parameters is correct):
<working_dir> - working directory used for storing in-transit files
<tmn_host> - base URL for tenant management node of Cloud Integration (excluding the https:// prefix)
<cpi_user> - user ID for Cloud Integration
<cpi_password> - password for above user ID
<artifact_id> - ID of integration artifact
<artifact_name> - name of integration artifact
<package_id> - ID of Integration Package
<package_name> - name of Integration Package
<git_src_dir> - directory containing contents of integration artifact

Optional parameters (position must always be in front of mandatory parameters):
--param_file=<path_to_file> - use to a different parameters.prop file instead of the default in src/main/resources/
--manifest_file=<path_to_file> - use to a different MANIFEST.MF file instead of the default in META-INF/
--logcfgfile=<path_to_file> - use a different log4j2.xml configuration file
--debug - display script debugging logs, e.g. detailed comparison output of artifacts

NOTE: Encapsulate values in double quotes ("") if there are space characters in them
```

#### Example
```bash
/usr/bin/update_designtime_artifact.sh /tmp ***.hana.ondemand.com mycpiuser mycpipassword GroovyXMLTransformation "Groovy XML Transformation" FlashPipeDemo "FlashPipe Demo" "FlashPipe Demo/Groovy XML Transformation"
```

### 2. deploy_runtime_artifact.sh
This script is used to deploy a Cloud Integration designtime artifact to the runtime.


#### Usage and parameters list
```bash
/usr/bin/deploy.sh [--logcfgfile=<pathtofile>] <artifact_id> <tmn_host> <cpi_user> <cpi_password>

Mandatory parameters (ensure order of parameters is correct):
<artifact_id> - ID of integration artifact
<tmn_host> - base URL for tenant management node of Cloud Integration (excluding the https:// prefix)
<cpi_user> - user ID for Cloud Integration
<cpi_password> - password for above user ID

Optional parameter (position must always be in front of mandatory parameters):
--logcfgfile=<path_to_file> - use a different log4j2.xml configuration file
```

#### Example
```bash
/usr/bin/deploy_runtime_artifact.sh GroovyXMLTransformation ***.hana.ondemand.com mycpiuser mycpipassword
```