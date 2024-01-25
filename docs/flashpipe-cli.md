# FlashPipe CLI

_FlashPipe_ provides a fully functional CLI with the following commands to interact with SAP Integration Suite.
- **[update artifact](#1-update-artifact)**
- **[update package](#2-update-package)**
- **[deploy](#3-deploy)**
- **[sync](#4-sync)**
- **[sync apim](#5-sync-apim)**
- **[snapshot](#6-snapshot)**


These commands perform the _magic_ that significantly simplifies the steps required to execute the build and deploy steps in a CI/CD pipeline.

The following section describes the usage of the commands. Input values can be passed using command line flags, environment variables or a config file (`flashpipe.yaml`).

With the support of [Viper](https://github.com/spf13/viper), each CLI flag can be substituted with a corresponding environment variable with the following rule - add "FLASHPIPE_" prefix, change name to upper case, and replace dashes with underscores. For example:

> --artifact-id >>> FLASHPIPE_ARTIFACT_ID

### Global flags
The following global flags and corresponding environment variables are available for all commands.

| CLI flag name      | Environment variable name    | Mandatory                     | Description                                                                               |
|--------------------|------------------------------|-------------------------------|-------------------------------------------------------------------------------------------|
| tmn-host           | FLASHPIPE_TMN_HOST           | Yes                           | Host for tenant management node of Cloud Integration or API Management excluding https:// |
| tmn-userid         | FLASHPIPE_TMN_USERID         | Yes (if OAuth Host is empty)  | User ID for Basic Auth                                                                    |
| tmn-password       | FLASHPIPE_TMN_PASSWORD       | Yes (if OAuth Host is empty)  | Password for Basic Auth                                                                   |
| oauth-host         | FLASHPIPE_OAUTH_HOST         | No                            | Host for OAuth token server excluding https://                                            |
| oauth-clientid     | FLASHPIPE_OAUTH_CLIENTID     | Yes (if OAuth Host is filled) | Client ID for using OAuth                                                                 |
| oauth-clientsecret | FLASHPIPE_OAUTH_CLIENTSECRET | Yes (if OAuth Host is filled) | Client Secret for using OAuth                                                             |
| oauth-path         | FLASHPIPE_OAUTH_PATH         | No                            | Path for OAuth token server (default "/oauth/token")                                      |
| debug              | FLASHPIPE_DEBUG              | No                            | Show debug logs                                                                           |
| config             | FLASHPIPE_CONFIG             | No                            | config file (default is $HOME/flashpipe.yaml)                                             |

### 1. update artifact
This command is used to create/update a Cloud Integration designtime artifact on the tenant. It provides the following functionalities:
- check existence of artifact to determine if it needs to be created or updated
- create Integration Package (if it does not exist) to store the artifact
- compare contents of artifact in Git repository against tenant to determine if artifact in tenant needs to be updated
- use different `parameters.prop` files to handle different configuration values when deploying multiple copies of artifact to same/different tenants
- create/update designtime artifact
- handle conversion of script collection references (for deployment of multiple copies in same tenant/different tenants)


#### Usage
```bash
flashpipe update artifact -h

Create or update artifacts on the
SAP Integration Suite tenant.

Usage:
  flashpipe update artifact [flags]

Flags:
      --artifact-id string             ID of artifact
      --artifact-name string           Name of artifact. Defaults to artifact-id value when not provided
      --artifact-type string           Artifact type. Allowed values: Integration, MessageMapping, ScriptCollection, ValueMapping (default "Integration")
      --dir-artifact string            Directory containing contents of designtime artifact
      --dir-work string                Working directory for in-transit files (default "/tmp")
      --file-manifest string           Use a different MANIFEST.MF file instead of the default in META-INF/
      --file-param string              Use a different parameters.prop file instead of the default in src/main/resources/ 
  -h, --help                           help for artifact
      --package-id string              ID of Integration Package
      --package-name string            Name of Integration Package. Defaults to package-id value when not provided
      --script-collection-map string   Comma-separated source-target ID pairs for converting script collection references during create/update

Global Flags:
      --config string               config file (default is $HOME/flashpipe.yaml)
      --debug                       Show debug logs
      --oauth-clientid string       Client ID for using OAuth
      --oauth-clientsecret string   Client Secret for using OAuth
      --oauth-host string           Host for OAuth token server excluding https:// 
      --oauth-path string           Path for OAuth token server (default "/oauth/token")
      --tmn-host string             Host for tenant management node of Cloud Integration excluding https://
      --tmn-password string         Password for Basic Auth
      --tmn-userid string           User ID for Basic Auth

NOTE: Encapsulate values in double quotes ("") if there are space characters in them
```

#### CLI flags and environment variables list
The following is the list of flags for the `update artifact` command and their corresponding environment variable name.

| CLI flag name         | Environment variable name       | Mandatory |
|-----------------------|---------------------------------|-----------|
| artifact-id           | FLASHPIPE_ARTIFACT_ID           | Yes       |
| artifact-name         | FLASHPIPE_ARTIFACT_NAME         | No        |
| package-id            | FLASHPIPE_PACKAGE_ID            | Yes       |
| package-name          | FLASHPIPE_PACKAGE_NAME          | No        |
| dir-artifact          | FLASHPIPE_DIR_ARTIFACT          | Yes       |
| artifact-type         | FLASHPIPE_ARTIFACT_TYPE         | No        |
| file-param            | FLASHPIPE_FILE_PARAM            | No        |
| file-manifest         | FLASHPIPE_FILE_MANIFEST         | No        |
| dir-work              | FLASHPIPE_DIR_WORK              | No        |
| script-collection-map | FLASHPIPE_SCRIPT_COLLECTION_MAP | No        |


#### Example (Basic Auth with CLI flags)
```bash
flashpipe update artifact --tmn-host ***.hana.ondemand.com --tmn-userid <userid> --tmn-password <password> --artifact-id GroovyXMLTransformation --artifact-name "Groovy XML Transformation" --package-id FlashPipeDemo --package-name "FlashPipe Demo" --dir-artifact "FlashPipe Demo/Groovy XML Transformation"
```

#### Example (OAuth with environment variables)
```bash
flashpipe update artifact

Environment variables set before call:
    FLASHPIPE_TMN_HOST: ***.hana.ondemand.com
    FLASHPIPE_OAUTH_HOST: ***.authentication.<region>.hana.ondemand.com
    FLASHPIPE_OAUTH_CLIENTID: <clientid>
    FLASHPIPE_OAUTH_CLIENTSECRET: <clientsecret>
    FLASHPIPE_ARTIFACT_ID: GroovyXMLTransformation
    FLASHPIPE_ARTIFACT_NAME: "Groovy XML Transformation"
    FLASHPIPE_PACKAGE_ID: FlashPipeDemo
    FLASHPIPE_PACKAGE_NAME: "FlashPipe Demo"
    FLASHPIPE_DIR_ARTIFACT: "FlashPipe Demo/Groovy XML Transformation"
```


### 2. update package
This command is used to create/update a Cloud Integration `integration package` to the tenant. It provides the following functionalities:
- check existence of package to determine if it needs to be created or updated
- compare contents of package in Git repository against tenant to determine if package in tenant needs to be updated
- create/update integration package


#### Usage
```bash
flashpipe update package -h

Create or update integration package on the
SAP Integration Suite tenant.

Usage:
  flashpipe update package [flags]

Aliases:
  package, pkg

Flags:
  -h, --help                  help for package
      --package-file string   Path to location of package file

Global Flags:
      --config string               config file (default is $HOME/flashpipe.yaml)
      --debug                       Show debug logs
      --oauth-clientid string       Client ID for using OAuth
      --oauth-clientsecret string   Client Secret for using OAuth
      --oauth-host string           Host for OAuth token server excluding https:// 
      --oauth-path string           Path for OAuth token server (default "/oauth/token")
      --tmn-host string             Host for tenant management node of Cloud Integration excluding https://
      --tmn-password string         Password for Basic Auth
      --tmn-userid string           User ID for Basic Auth
```

#### CLI flags and environment variables list
The following is the list of flags for the `update package` command and their corresponding environment variable name.

| CLI flag name         | Environment variable name | Mandatory |
|-----------------------|---------------------------|-----------|
| package-file          | FLASHPIPE_PACKAGE_FILE    | Yes       |

#### Example (Basic Auth with CLI flags)
```bash
flashpipe update package --tmn-host ***.hana.ondemand.com --tmn-userid <userid> --tmn-password <password> --package-file "<path_to_file>/FlashPipeDemo.json"
```

#### Example (OAuth with environment variables)
```bash
flashpipe update package

Environment variables set before call:
    FLASHPIPE_TMN_HOST: ***.hana.ondemand.com
    FLASHPIPE_OAUTH_HOST: ***.authentication.<region>.hana.ondemand.com
    FLASHPIPE_OAUTH_CLIENTID: <clientid>
    FLASHPIPE_OAUTH_CLIENTSECRET: <clientsecret>
    FLASHPIPE_PACKAGE_FILE: "<path_to_file>/FlashPipeDemo.json"
```

### 3. deploy
This command is used to deploy Cloud Integration designtime artifact(s) to the runtime. It can compare the version of the designtime artifact against the runtime artifact before executing deployment if there are differences.


#### Usage
```bash
flashpipe deploy -h

Deploy artifact from designtime to
runtime of SAP Integration Suite tenant.

Usage:
  flashpipe deploy [flags]

Flags:
      --artifact-ids string    Comma separated list of artifact IDs
      --artifact-type string   Artifact type. Allowed values: Integration, MessageMapping, ScriptCollection, ValueMapping (default "Integration")
      --compare-versions       Perform version comparison of design time against runtime before deployment (default true)
      --delay-length int       Delay (in seconds) between each check of artifact deployment status (default 30)
  -h, --help                   help for deploy
      --max-check-limit int    Max number of times to check for artifact deployment status (default 10)

Global Flags:
      --config string               config file (default is $HOME/flashpipe.yaml)
      --debug                       Show debug logs
      --oauth-clientid string       Client ID for using OAuth
      --oauth-clientsecret string   Client Secret for using OAuth
      --oauth-host string           Host for OAuth token server excluding https:// 
      --oauth-path string           Path for OAuth token server (default "/oauth/token")
      --tmn-host string             Host for tenant management node of Cloud Integration excluding https://
      --tmn-password string         Password for Basic Auth
      --tmn-userid string           User ID for Basic Auth
```

#### CLI flags and environment variables list
The following is the list of flags for the `deploy` command and their corresponding environment variable name.

| CLI flag name    | Environment variable name  | Mandatory |
|------------------|----------------------------|-----------|
| artifact-ids     | FLASHPIPE_ARTIFACT_IDS     | Yes       |
| artifact-type    | FLASHPIPE_ARTIFACT_TYPE    | No        |
| compare-versions | FLASHPIPE_COMPARE_VERSIONS | No        |
| delay-length     | FLASHPIPE_DELAY_LENGTH     | No        |
| max-check-limit  | FLASHPIPE_MAX_CHECK_LIMIT  | No        |

#### Example (Basic Auth with CLI flags)
```bash
flashpipe deploy --tmn-host ***.hana.ondemand.com --tmn-userid <userid> --tmn-password <password> --artifact-ids GroovyXMLTransformation
```

#### Example (OAuth with environment variables)
```bash
flashpipe deploy

Environment variables set before call:
    FLASHPIPE_TMN_HOST: ***.hana.ondemand.com
    FLASHPIPE_OAUTH_HOST: ***.authentication.<region>.hana.ondemand.com
    FLASHPIPE_OAUTH_CLIENTID: <clientid>
    FLASHPIPE_OAUTH_CLIENTSECRET: <clientsecret>
    FLASHPIPE_ARTIFACT_IDS: GroovyXMLTransformation
```

### 4. sync
This command is used to sync Cloud Integration designtime artifacts and integration package details (optional) between a tenant and a Git repository. It will compare any differences (new, deleted, changed) in files between tenant and the Git repository before synchronising them.


#### Usage
```bash
flashpipe sync -h

Synchronise designtime artifacts between SAP Integration Suite
tenant and a Git repository.

Usage:
  flashpipe sync [flags]

Flags:
      --dir-artifacts string           Directory containing contents of artifacts
      --dir-git-repo string            Directory of Git repository
      --dir-naming-type string         Name artifact directory by ID or Name. Allowed values: ID, NAME (default "ID")
      --dir-work string                Working directory for in-transit files (default "/tmp")
      --draft-handling string          Handling when artifact is in draft version. Allowed values: SKIP, ADD, ERROR (default "SKIP")
      --git-commit-email string        Email used in commit (default "41898282+github-actions[bot]@users.noreply.github.com")
      --git-commit-msg string          Message used in commit (default "Sync repo from tenant")
      --git-commit-user string         User used in commit (default "github-actions[bot]")
      --git-skip-commit                Skip committing changes to Git repository
  -h, --help                           help for sync
      --ids-exclude string             List of excluded artifact IDs
      --ids-include string             List of included artifact IDs
      --package-id string              ID of Integration Package
      --script-collection-map string   Comma-separated source-target ID pairs for converting script collection references during sync 
      --sync-package-details           Sync details of Integration Package
      --target                         Target of sync. Allowed values: git, tenant (default "git")

Global Flags:
      --config string               config file (default is $HOME/flashpipe.yaml)
      --debug                       Show debug logs
      --oauth-clientid string       Client ID for using OAuth
      --oauth-clientsecret string   Client Secret for using OAuth
      --oauth-host string           Host for OAuth token server excluding https:// 
      --oauth-path string           Path for OAuth token server (default "/oauth/token")
      --tmn-host string             Host for tenant management node of Cloud Integration excluding https://
      --tmn-password string         Password for Basic Auth
      --tmn-userid string           User ID for Basic Auth
```

#### CLI flags and environment variables list
The following is the list of flags for the `sync` command and their corresponding environment variable name. The fourth column indicates whether the flag is valid for the specific value of --target.

| CLI flag name         | Environment variable name       | Mandatory | Applicable for value of --target |
|-----------------------|---------------------------------|-----------|----------------------------------|
| package-id            | FLASHPIPE_PACKAGE_ID            | Yes       | git, tenant                      |
| dir-git-repo          | FLASHPIPE_DIR_GIT_REPO          | Yes       | git, tenant                      |
| dir-artifacts         | FLASHPIPE_DIR_ARTIFACTS         | No        | git, tenant                      |
| target                | FLASHPIPE_TARGET                | No        | git, tenant                      |
| dir-naming-type       | FLASHPIPE_DIR_NAMING_TYPE       | No        | git                              |
| draft-handling        | FLASHPIPE_DRAFT_HANDLING        | No        | git                              |
| ids-include           | FLASHPIPE_IDS_INCLUDE           | No        | git, tenant                      |
| ids-exclude           | FLASHPIPE_IDS_EXCLUDE           | No        | git, tenant                      |
| git-commit-msg        | FLASHPIPE_GIT_COMMIT_MSG        | No        | git                              |
| git-commit-user       | FLASHPIPE_GIT_COMMIT_USER       | No        | git                              |
| git-commit-email      | FLASHPIPE_GIT_COMMIT_EMAIL      | No        | git                              |
| git-skip-commit       | FLASHPIPE_GIT_SKIP_COMMIT       | No        | git                              |
| script-collection-map | FLASHPIPE_SCRIPT_COLLECTION_MAP | No        | git                              |
| sync-package-details  | FLASHPIPE_SYNC_PACKAGE_DETAILS  | No        | git                              |
| dir-work              | FLASHPIPE_DIR_WORK              | No        | git, tenant                      |

#### Example (Basic Auth with CLI flags)
```bash
flashpipe sync --tmn-host ***.hana.ondemand.com --tmn-userid <userid> --tmn-password <password> --package-id FlashPipeDemo --dir-git-repo "FlashPipe Demo"
```

#### Example (OAuth with environment variables)
```bash
flashpipe sync

Environment variables set before call:
    FLASHPIPE_TMN_HOST: ***.hana.ondemand.com
    FLASHPIPE_OAUTH_HOST: ***.authentication.<region>.hana.ondemand.com
    FLASHPIPE_OAUTH_CLIENTID: <clientid>
    FLASHPIPE_OAUTH_CLIENTSECRET: <clientsecret>
    FLASHPIPE_PACKAGE_ID:  FlashPipeDemo
    FLASHPIPE_DIR_GIT_REPO: "FlashPipe Demo"
    FLASHPIPE_DIR_ARTIFACTS: "FlashPipe Demo/Contents"
    FLASHPIPE_SYNC_PACKAGE_DETAILS: true
```

### 5. sync apim
This command is used to sync API Management artifacts between a tenant and a Git repository. It will compare any differences (new, deleted, changed) in files between tenant and the Git repository before synchronising them.
- dependent artifacts of the API Proxy are included like API Provider, Key Value Maps

#### Usage
```bash
flashpipe sync apim -h

Synchronise API Management artifacts between SAP Integration Suite
tenant and a Git repository.

Usage:
  flashpipe sync apim [flags]

Flags:
      --dir-artifacts string           Directory containing contents of artifacts
      --dir-git-repo string            Directory of Git repository
      --dir-work string                Working directory for in-transit files (default "/tmp")
      --git-commit-email string        Email used in commit (default "41898282+github-actions[bot]@users.noreply.github.com")
      --git-commit-msg string          Message used in commit (default "Sync repo from tenant")
      --git-commit-user string         User used in commit (default "github-actions[bot]")
      --git-skip-commit                Skip committing changes to Git repository
  -h, --help                           help for sync
      --ids-exclude string             List of excluded artifact IDs
      --ids-include string             List of included artifact IDs
      --target                         Target of sync. Allowed values: git, tenant (default "git")

Global Flags:
      --config string               config file (default is $HOME/flashpipe.yaml)
      --debug                       Show debug logs
      --oauth-clientid string       Client ID for using OAuth
      --oauth-clientsecret string   Client Secret for using OAuth
      --oauth-host string           Host for OAuth token server excluding https:// 
      --oauth-path string           Path for OAuth token server (default "/oauth/token")
      --tmn-host string             Host for API Portal for API Management excluding https://
```

#### CLI flags and environment variables list
The following is the list of flags for the `sync apim` command and their corresponding environment variable name. The fourth column indicates whether the flag is valid for the specific value of --target.

| CLI flag name         | Environment variable name       | Mandatory | Applicable for value of --target |
|-----------------------|---------------------------------|-----------|----------------------------------|
| dir-git-repo          | FLASHPIPE_DIR_GIT_REPO          | Yes       | git, tenant                      |
| dir-artifacts         | FLASHPIPE_DIR_ARTIFACTS         | No        | git, tenant                      |
| target                | FLASHPIPE_TARGET                | No        | git, tenant                      |
| ids-include           | FLASHPIPE_IDS_INCLUDE           | No        | git, tenant                      |
| ids-exclude           | FLASHPIPE_IDS_EXCLUDE           | No        | git, tenant                      |
| git-commit-msg        | FLASHPIPE_GIT_COMMIT_MSG        | No        | git                              |
| git-commit-user       | FLASHPIPE_GIT_COMMIT_USER       | No        | git                              |
| git-commit-email      | FLASHPIPE_GIT_COMMIT_EMAIL      | No        | git                              |
| git-skip-commit       | FLASHPIPE_GIT_SKIP_COMMIT       | No        | git                              |
| dir-work              | FLASHPIPE_DIR_WORK              | No        | git, tenant                      |

#### Example (OAuth with CLI flags)
```bash
flashpipe sync apim --tmn-host ***.hana.ondemand.com --oauth-host ***.authentication.<region>.hana.ondemand.com --oauth-clientid <clientid> --oauth-clientsecret <clientsecret> --dir-git-repo "FlashPipe APIM Demo"
```

#### Example (OAuth with environment variables)
```bash
flashpipe sync apim

Environment variables set before call:
    FLASHPIPE_TMN_HOST: ***.hana.ondemand.com
    FLASHPIPE_OAUTH_HOST: ***.authentication.<region>.hana.ondemand.com
    FLASHPIPE_OAUTH_CLIENTID: <clientid>
    FLASHPIPE_OAUTH_CLIENTSECRET: <clientsecret>
    FLASHPIPE_DIR_GIT_REPO: "FlashPipe APIM Demo"
    FLASHPIPE_DIR_ARTIFACTS: "FlashPipe APIM Demo/Contents"
```

### 6. snapshot
This command is used to capture a snapshot of the Cloud Integration tenant's artifacts and integration package details (optional) to a Git repository. It will compare any differences (new, deleted, changed) in files from tenant and commit/push to the Git repository.


#### Usage
```bash
flashpipe snapshot -h

Snapshot all editable integration packages from SAP Integration Suite
tenant to a Git repository.

Usage:
  flashpipe snapshot [flags]

Flags:
      --dir-git-repo string       Directory of Git repository containing contents of artifacts (grouped into packages)
      --dir-work string           Working directory for in-transit files (default "/tmp")
      --draft-handling string     Handling when artifact is in draft version. Allowed values: SKIP, ADD, ERROR (default "SKIP")
      --git-commit-email string   Email used in commit (default "41898282+github-actions[bot]@users.noreply.github.com")
      --git-commit-msg string     Message used in commit (default "Tenant snapshot of <current timestamp>")
      --git-commit-user string    User used in commit (default "github-actions[bot]")
      --git-skip-commit           Skip committing changes to Git repository
  -h, --help                      help for snapshot
      --sync-package-details      Sync details of Integration Packages

Global Flags:
      --config string               config file (default is $HOME/flashpipe.yaml)
      --debug                       Show debug logs
      --oauth-clientid string       Client ID for using OAuth
      --oauth-clientsecret string   Client Secret for using OAuth
      --oauth-host string           Host for OAuth token server excluding https:// 
      --oauth-path string           Path for OAuth token server (default "/oauth/token")
      --tmn-host string             Host for tenant management node of Cloud Integration excluding https://
      --tmn-password string         Password for Basic Auth
      --tmn-userid string           User ID for Basic Auth
```

#### CLI flags and environment variables list
The following is the list of flags for the `snapshot` command and their corresponding environment variable name.

| CLI flag name         | Environment variable name       | Mandatory |
|-----------------------|---------------------------------|-----------|
| dir-git-repo          | FLASHPIPE_DIR_GIT_REPO          | Yes       |
| draft-handling        | FLASHPIPE_DRAFT_HANDLING        | No        |
| git-commit-msg        | FLASHPIPE_GIT_COMMIT_MSG        | No        |
| git-commit-user       | FLASHPIPE_GIT_COMMIT_USER       | No        |
| git-commit-email      | FLASHPIPE_GIT_COMMIT_EMAIL      | No        |
| git-skip-commit       | FLASHPIPE_GIT_SKIP_COMMIT       | No        |
| sync-package-details  | FLASHPIPE_SYNC_PACKAGE_DETAILS  | No        |
| dir-work              | FLASHPIPE_DIR_WORK              | No        |

#### Example (Basic Auth with CLI flags)
```bash
flashpipe snapshot --tmn-host ***.hana.ondemand.com --tmn-userid <userid> --tmn-password <password> --dir-git-repo "TrialTenant"
```

#### Example (OAuth with environment variables)
```bash
flashpipe snapshot

Environment variables set before call:
    FLASHPIPE_TMN_HOST: ***.hana.ondemand.com
    FLASHPIPE_OAUTH_HOST: ***.authentication.<region>.hana.ondemand.com
    FLASHPIPE_OAUTH_CLIENTID: <clientid>
    FLASHPIPE_OAUTH_CLIENTSECRET: <clientsecret>
    FLASHPIPE_DIR_GIT_REPO: "TrialTenant"
```
