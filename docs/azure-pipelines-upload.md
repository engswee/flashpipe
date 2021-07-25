# Upload/Deploy with FlashPipe on Azure Pipelines
The page describes the steps to set up _FlashPipe_ on [Azure Pipelines](https://azure.microsoft.com/en-us/services/devops/pipelines/).

**Note**: [GitHub repository syncing from tenant](github-actions-sync.md) can also be used in place of steps 1 and 2.

### 1. Download and extract content of Integration Flow
Download the content of the Integration Flow from the Cloud Integration tenant.
![Download](images/setup/01a_download_iflow.png)

Extract the content of the downloaded ZIP file
![Content](images/setup/01b_iflow_contents.png)

### 2. Add content to Git repository
Add the contents to a new or existing Git repository.
![Git](images/setup/02a_add_to_git.png)

### 3. Add Maven POM for unit testing [Optional]
If you intend to execute unit testing using Maven, add a Maven POM file (`pom.xml`) to the Git repository with the appropriate content.

_FlashPipe_'s Maven repository comes loaded with the following libraries (and any dependencies), so you can gain advantage of faster execution time by running in offline mode `mvn test -o`.
- org.codehaus.groovy:groovy-all:2.4.21
- org.spockframework:spock-core:1.3-groovy-2.4
- org.apache.camel:camel-core:2.24.2
- org.apache.httpcomponents.client5:httpclient5:5.0.4
- org.apache.logging.log4j:log4j-api:2.14.1
- org.apache.logging.log4j:log4j-core:2.14.1
- org.apache.logging.log4j:log4j-slf4j-impl:2.14.1
- net.bytebuddy:byte-buddy:1.11.0

For multiple integration packages/artifacts, the repository can be structured as a Maven multi-module project. An example can be found [here](https://github.com/engswee/flashpipe-demo/tree/azure-pipelines).

### 4. Add pipeline configuration YAML
Add a [pipeline configuration YAML file](https://docs.microsoft.com/en-us/azure/devops/pipelines/get-started/pipelines-get-started?view=azure-devops#define-pipelines-using-yaml-syntax) (`azure-pipelines.yml`) in the root directory of the Git repository.

#### Template YAML with steps to create/update and deploy one integration artifact
```yaml
trigger:
  - <branch_name>

pool:
  vmImage: 'ubuntu-latest'

variables:
  - group: <variable_group_name>

resources:
  containers:
    - container: flashpipe
      image: engswee/flashpipe:<flashpipe_version>

jobs:
  - job: build
    container: flashpipe
    steps:
      # Upload/Update design time
      - bash: /usr/bin/update_designtime_artifact.sh
        env:
          HOST_TMN: $(dev-host-tmn)
          BASIC_USERID: $(dev-user) # When using Basic authentication
          BASIC_PASSWORD: $(dev-password) # When using Basic authentication
          HOST_OAUTH: $(dev-oauth-host) # When using OAuth authentication
          HOST_OAUTH_PATH: <oauth_path> # Optional - set to /oauth2/api/v1/token for Neo environments
          OAUTH_CLIENTID: $(dev-client-id) # When using OAuth authentication
          OAUTH_CLIENTSECRET: $(dev-client-secret) # When using OAuth authentication
          IFLOW_ID: <iflow_id>
          IFLOW_NAME: <iflow_name>
          PACKAGE_ID: <package_id>
          PACKAGE_NAME: <package_name>
          GIT_SRC_DIR: <git_src_dir>
          PARAM_FILE: <param_file> # Optional
          MANIFEST_FILE: <manifest_file> # Optional
          WORK_DIR: <working_directory> # Optional
      # Deploy runtime
      - bash: /usr/bin/deploy_runtime_artifact.sh
        env:
          HOST_TMN: $(dev-host-tmn)
          BASIC_USERID: $(dev-user) # When using Basic authentication
          BASIC_PASSWORD: $(dev-password) # When using Basic authentication
          HOST_OAUTH: $(dev-oauth-host) # When using OAuth authentication
          HOST_OAUTH_PATH: <oauth_path> # Optional - set to /oauth2/api/v1/token for Neo environments
          OAUTH_CLIENTID: $(dev-client-id) # When using OAuth authentication
          OAUTH_CLIENTSECRET: $(dev-client-secret) # When using OAuth authentication
          IFLOW_ID: <iflow_id>
          DELAY_LENGTH: <delay_in_seconds> # Optional
          MAX_CHECK_LIMIT: <max_check_limit> # Optional
```
Where:
- `<branch_name>` - branch name of Git repository that will automatically trigger pipeline
- `<variable_group_name>` - name of Azure Pipeline variable group that stores environment variables for access to Cloud Integration tenant - `$(dev-host-tmn), $(dev-user), $(dev-password)`. Further explanation in step 6
- `<flashpipe_version>` - version of _FlashPipe_

**Note**: Environment variables are mapped to the script's execution environment using the `env:` keyword.

#### Example (using OAuth authentication for Cloud Foundry)

```yaml
trigger:
  - main

pool:
  vmImage: 'ubuntu-latest'

variables:
  - group: cpi-dev

resources:
  containers:
    - container: flashpipe
      image: engswee/flashpipe:2.2.2-lite

jobs:
  - job: build
    container: flashpipe
    steps:
      # Upload/Update design time
      - bash: /usr/bin/update_designtime_artifact.sh
        displayName: 'Update/Upload Groovy XML Transformation to design time'
        env:
          HOST_TMN: $(dev-host-tmn)
          HOST_OAUTH: $(dev-oauth-host)
          OAUTH_CLIENTID: $(dev-client-id)
          OAUTH_CLIENTSECRET: $(dev-client-secret)
          IFLOW_ID: GroovyXMLTransformation
          IFLOW_NAME: "Groovy XML Transformation"
          PACKAGE_ID: FlashPipeDemo
          PACKAGE_NAME: "FlashPipe Demo"
          GIT_SRC_DIR: "$(Build.SourcesDirectory)/FlashPipe Demo/Groovy XML Transformation"
      # Deploy runtime
      - bash: /usr/bin/deploy_runtime_artifact.sh
        displayName: 'Deploy Groovy XML Transformation to runtime'
        env:
          HOST_TMN: $(dev-host-tmn)
          HOST_OAUTH: $(dev-oauth-host)
          OAUTH_CLIENTID: $(dev-client-id)
          OAUTH_CLIENTSECRET: $(dev-client-secret)
          IFLOW_ID: GroovyXMLTransformation
```

For more advanced configuration with multiple artifacts and multiple environments, an example can be found [here](https://github.com/engswee/flashpipe-demo/blob/azure-pipelines/azure-pipelines.yml).

### 5. Create new project in Azure DevOps
![Project](images/setup/azure-pipelines/05a_azure_project.png)

### 6. Create Variable Group
Variables can be stored securely on Azure Pipelines using a [Variable Group](https://docs.microsoft.com/en-us/azure/devops/pipelines/library/variable-groups?view=azure-devops&tabs=yaml). These can then be passed to the pipeline steps as environment variables. For _FlashPipe_, we will use these to securely store the details to access the Cloud Integration tenant.

Create a new Variable Group under `Pipelines > Library`. Use the same name as defined in the `variables` section of the YAML, e.g. `cpi-dev`
![Library](images/setup/azure-pipelines/06a_library.png)

**Basic Authentication**

Add the following three variables in the group.
1. `dev-host-tmn` - base URL for tenant management node of Cloud Integration (excluding the https:// prefix)
2. `dev-user` - user ID for Cloud Integration
3. `dev-password` - password for above user ID
   ![Variable Group](images/setup/azure-pipelines/06b_variable_group_basic.png)

**OAuth Authentication**

Add the following four variables in the group. Refer to [OAuth client setup page](oauth_client.md) for details on setting up the OAuth client for usage with _FlashPipe_.
1. `dev-host-tmn` - base URL for tenant management node of Cloud Integration (excluding the https:// prefix)
2. `dev-oauth-host` - host name for OAuth authentication server
3. `dev-client-id` - OAuth client ID
4. `dev-client-secret` - OAuth client secret
   ![Variable Group](images/setup/azure-pipelines/06c_variable_group_oauth.png)

**Note**: For the password and client secret (and optionally the user ID), it can be stored securely as a secret instead of plain text by clicking the padlock button on its right.

### 7. Create new pipeline based on Git repository
Next, move on to create a new pipeline in the Azure DevOps project.
![Pipeline](images/setup/azure-pipelines/07a_pipeline.png)

Select the Git repository to be used in the pipeline.
![Select Repo](images/setup/azure-pipelines/07b_select_repo.png)

Since the pipeline YAML file is already created in the repository, it will be loaded. Review it and then select `Run` to execute the pipeline.
![Review Run](images/setup/azure-pipelines/07c_review_run.png)

### 8. Check pipeline run
Once the run is triggered, you can monitor its execution and job logs.

**Note**: On the first run of the pipeline, you may be asked to approve access to the Variable Group from the pipeline.

Upon completion of the run, you can review the logs, and also check the artifact (designtime and runtime) in the Cloud Integration tenant.
![Monitor](images/setup/azure-pipelines/08a_job_run.png)