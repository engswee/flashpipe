# Setting Up FlashPipe on GitHub Actions
The page describes the steps to set up _FlashPipe_ on [GitHub Actions](https://github.com/features/actions).

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

### 4. Add GitHub Actions workflow YAML
Add a [GitHub Actions workflow YAML file](https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions) (`<workflow-name>.yml`) in the `.github/workflows` directory of the Git repository.

#### Template YAML with steps to create/update and deploy one integration artifact
<script src="https://gist.github.com/engswee/06a528a4dbd7278e8a1020dfda5bd9b6.js"></script>

Where:
- `<branch_name>` - branch name of Git repository that will automatically trigger pipeline
- `<flashpipe_version>` - version of _FlashPipe_
- `secrets.<name>` - Sensitive information are stored as encrypted secrets in GitHub and accessed using the `secrets` context. Further explanation in step 5

**Note**: Environment variables are mapped to the script's execution environment using the `env:` keyword. For variables that are dynamic expressions based on other variables, these needs to be stored into the `$GITHUB_ENV` variable prior to the script execution. An example shown above is `$GIT_DIR` which requires base path from `$GITHUB_WORKSPACE`.

#### Example (using OAuth authentication)
<script src="https://gist.github.com/engswee/9de198d84650c08b7cdae4e7c08e1bcd.js"></script>

### 5. Create secrets in GitHub repository
Sensitive information can be stored securely on GitHub using [encrypted secrets](https://docs.github.com/en/actions/reference/encrypted-secrets). These can then be passed to the pipeline steps as environment variables. For _FlashPipe_, we will use these to securely store the details to access the Cloud Integration tenant.

In the GitHub repository, go to `Settings` > `Secrets` to create new repository secrets as shown below.
![Secrets Setting](images/setup/github-actions/05a_secrets.png)

**Basic Authentication**

Create the following repository secrets.
1. `DEV_USER_ID` - user ID for Cloud Integration
2. `DEV_PASSWORD` - password for above user ID
   ![Basic Secrets](images/setup/github-actions/05b_basic_secrets.png)

**OAuth Authentication**

Create the following repository secrets.
1. `DEV_CLIENT_ID` - OAuth client ID
2. `DEV_CLIENT_SECRET` - OAuth client secret
   ![OAuth Secrets](images/setup/github-actions/05c_oauth_secrets.png)

**Note**: GitHub does not provide functionality to store unencrypted plain text variables, which would be useful for values like the base URLs. Optionally, these can be stored as encrypted secrets instead of being hardcoded in the YAML configuration file.

### 6. Commit/push the workflow YAML, and check pipeline run
Once all is in place, commit/push the workflow YAML. This will automatically trigger the workflow to be executed, you can monitor its execution and job logs. Go to `Actions` to view the workflows.
![Monitor](images/setup/github-actions/06a_action_workflow.png)

Upon completion of the run, you can review the logs, and also check the artifact (designtime and runtime) in the Cloud Integration tenant.
![Monitor](images/setup/github-actions/06b_action_logs.png)