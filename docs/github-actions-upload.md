# Upload/Deploy with FlashPipe on GitHub Actions
The page describes the steps to set up _FlashPipe_ on [GitHub Actions](https://github.com/features/actions).

**Note**: [GitHub repository syncing from tenant](github-actions-sync.md) can also be used in place of steps 1 and 2.

### 1. Download and extract content of Integration Flow
Download the content of the Integration Flow from the Cloud Integration tenant.
![Download](images/setup/01a_download_iflow.png)

Extract the content of the downloaded ZIP file
![Content](images/setup/01b_iflow_contents.png)

### 2. Add content to Git repository
Add the contents to a new or existing Git repository.
![Git](images/setup/02a_add_to_git.png)

### 3. Add GitHub Actions workflow YAML
Add a [GitHub Actions workflow YAML file](https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions) (`<workflow-name>.yml`) in the `.github/workflows` directory of the Git repository.

#### Template YAML with steps to create/update and deploy one integration artifact
<script src="https://gist.github.com/engswee/b040f9c520c42ed8eb3307ec29c1e77a.js"></script>

```yaml
# https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions
name: <workflow-name>
on:
  push:
    branches:
      - <branch_name>

jobs:
  build:
    runs-on: ubuntu-latest
    container:
      image: engswee/flashpipe:<flashpipe_version>
    env:
      FLASHPIPE_TMN_HOST: <base URL of tenant management node>
      FLASHPIPE_TMN_USERID: ${{ secrets.DEV_USER_ID }} # When using Basic authentication
      FLASHPIPE_TMN_PASSWORD: ${{ secrets.DEV_PASSWORD }} # When using Basic authentication
      FLASHPIPE_OAUTH_HOST: <base URL of OAuth token server> # When using OAuth authentication
      FLASHPIPE_OAUTH_CLIENTID: ${{ secrets.DEV_CLIENT_ID }} # When using OAuth authentication
      FLASHPIPE_OAUTH_CLIENTSECRET: ${{ secrets.DEV_CLIENT_SECRET }} # When using OAuth authentication
      FLASHPIPE_OAUTH_PATH: <oauth_path> # Optional
    steps:
      - uses: actions/checkout@v4
      # Upload/Update design time artifact
      - name: 'Update/Upload artifact to design time'
        run: flashpipe update artifact
        shell: bash
        env:
          FLASHPIPE_ARTIFACT_ID: <artifact_id>
          FLASHPIPE_ARTIFACT_NAME: <artifact_name>
          FLASHPIPE_PACKAGE_ID: <package_id>
          FLASHPIPE_PACKAGE_NAME: <package_name>
          FLASHPIPE_DIR_ARTIFACT: ${{ github.workspace }}/<path_to_artifact_dir> # Optional
          FLASHPIPE_FILE_PARAM: ${{ github.workspace }}/<path_to_param_file> # Optional
          FLASHPIPE_FILE_MANIFEST: ${{ github.workspace }}/<path_to_manifest_file> # Optional
          FLASHPIPE_DIR_WORK: <working_directory> # Optional
          FLASHPIPE_SCRIPT_COLLECTION_MAP: <comma_separated_source/target_pairs> # Optional
          FLASHPIPE_ARTIFACT_TYPE: <artifact_type> # Optional
          FLASHPIPE_PACKAGE_FILE: <package_file> # Optional
      # Deploy to runtime
      - name: 'Deploy artifact to runtime'
        run: flashpipe deploy
        shell: bash
        env:
          FLASHPIPE_ARTIFACT_IDS: <iflow_id>
          FLASHPIPE_DELAY_LENGTH: <delay_in_seconds> # Optional
          FLASHPIPE_MAX_CHECK_LIMIT: <max_check_limit> # Optional          
          FLASHPIPE_COMPARE_VERSIONS: <compare_versions> # Optional
          FLASHPIPE_ARTIFACT_TYPE: <artifact_type> # Optional
```

Where:
- `<branch_name>` - branch name of Git repository that will automatically trigger pipeline
- `<flashpipe_version>` - version of _FlashPipe_
- `secrets.<name>` - Sensitive information are stored as encrypted secrets in GitHub and accessed using the `secrets` context. Further explanation in step 4

**Note**: Environment variables are mapped to the script's execution environment using the `env:` keyword. For variables that are dynamic expressions based on other variables, these needs to be stored into the `$GITHUB_ENV` variable prior to the script execution. An example shown above is `$GIT_SRC_DIR` which requires base path from `$GITHUB_WORKSPACE`.

#### Example (using OAuth authentication for Cloud Foundry)
<script src="https://gist.github.com/engswee/4f163729cdbda8eb7a56010a9ae37ac6.js"></script>

### 4. Create secrets in GitHub repository
Sensitive information can be stored securely on GitHub using [encrypted secrets](https://docs.github.com/en/actions/reference/encrypted-secrets). These can then be passed to the pipeline steps as environment variables. For _FlashPipe_, we will use these to securely store the details to access the Cloud Integration tenant.

In the GitHub repository, go to `Settings` > `Secrets` to create new repository secrets as shown below.
![Secrets Setting](images/setup/github-actions/05a_secrets.png)

**Basic Authentication**

Create the following repository secrets.
1. `DEV_USER_ID` - user ID for Cloud Integration
2. `DEV_PASSWORD` - password for above user ID
   ![Basic Secrets](images/setup/github-actions/05b_basic_secrets.png)

**OAuth Authentication**

Create the following repository secrets. Refer to [OAuth client setup page](oauth_client.md) for details on setting up the OAuth client for usage with _FlashPipe_.
1. `DEV_CLIENT_ID` - OAuth client ID
2. `DEV_CLIENT_SECRET` - OAuth client secret
   ![OAuth Secrets](images/setup/github-actions/05c_oauth_secrets.png)

**Note**: GitHub does not provide functionality to store unencrypted plain text variables, which would be useful for values like the base URLs. Optionally, these can be stored as encrypted secrets instead of being hardcoded in the YAML configuration file.

### 5. Commit/push the workflow YAML, and check pipeline run
Once all is in place, commit/push the workflow YAML. This will automatically trigger the workflow to be executed, you can monitor its execution and job logs. Go to `Actions` to view the workflows.
![Monitor](images/setup/github-actions/06a_action_workflow.png)

Upon completion of the run, you can review the logs, and also check the artifact (designtime and runtime) in the Cloud Integration tenant.
![Monitor](images/setup/github-actions/06b_action_logs.png)