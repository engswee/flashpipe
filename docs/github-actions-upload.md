# Upload/Deploy with FlashPipe on GitHub Actions
The page describes the steps to set up _FlashPipe_ on [GitHub Actions](https://github.com/features/actions).

**Note**: [GitHub repository syncing from tenant](github-actions-sync-to-git.md) can also be used in place of steps 1 and 2.

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

FlashPipe comes with companion GitHub Action [engswee/flashpipe-action](https://github.com/engswee/flashpipe-action) that simplifies usage in a workflow. The following actions are used in the workflow:
- [engswee/flashpipe-action/update/artifact@v1](https://github.com/engswee/flashpipe-action#update-artifact)
- [engswee/flashpipe-action/deploy@v1](https://github.com/engswee/flashpipe-action#deploy)

#### Template YAML with steps to create/update and deploy one integration artifact
[//]: # (Gist is used because inline YAML does not render ${{ variables }} correctly)
<script src="https://gist.github.com/engswee/b040f9c520c42ed8eb3307ec29c1e77a.js"></script>

Where:
- `<branch_name>` - branch name of Git repository that will automatically trigger pipeline
- `<flashpipe_version>` - version of _FlashPipe_
- `secrets.<name>` - Sensitive information are stored as encrypted secrets in GitHub and accessed using the `secrets` context. Further explanation in step 4

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

Note: GitHub provides functionality to store unencrypted plain text as `repository variables`. Optionally, values like base URLs can be stored as repository variables instead of being hardcoded in the YAML configuration file, and can then be access using [the `vars` context](https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/store-information-in-variables#using-the-vars-context-to-access-configuration-variable-values) in the configuration file.

### 5. Commit/push the workflow YAML, and check pipeline run
Once all is in place, commit/push the workflow YAML. This will automatically trigger the workflow to be executed, you can monitor its execution and job logs. Go to `Actions` to view the workflows.
![Monitor](images/setup/github-actions/06a_action_workflow.png)

Upon completion of the run, you can review the logs, and also check the artifact (designtime and runtime) in the Cloud Integration tenant.
![Monitor](images/setup/github-actions/06b_action_logs.png)