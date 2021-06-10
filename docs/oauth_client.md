# Setting Up OAuth Client for FlashPipe
This page describes the steps to set up an OAuth client for use with _FlashPipe_.

## Required Roles and Role-Templates
_FlashPipe_ relies heavily on access to Cloud Integration's public APIs. As such, it requires specific roles/role-templates in order to be able to access those APIs. Following are the tasks and corresponding roles/role-templates that are required.

Tasks | Role (Neo) | Role-Templates (Cloud Foundry)
------------ | ------------- | -------------
Create/edit design time artifacts | `WebToolingWorkspace.Write`, `WebToolingWorkspace.Read` | `WorkspacePackagesEdit`
Deploy artifacts to runtime | `NodeManager.deploycontent`, `GenerationAndBuild.generationandbuildcontent` | `WorkspaceArtifactsDeploy`
Monitor runtime artifacts | `IntegrationOperationServer.read`, `NodeManager.read` | `MonitoringDataRead`

## Creating an OAuth Client on Cloud Foundry

### 1. Logon to SAP BTP Cockpit
Access the relevant Cloud Foundry space on SAP BTP Cockpit.
![BTP](images/oauth-client/cf/01_btp_cf_space.png)

### 2. Create new service instance
In the space, navigate to the `Services > Instances` and create a new instance.
![CreateInstance](images/oauth-client/cf/02_create_instance.png)

### 3. Enter instance details
To access Cloud Integration APIs, we will enter the following details for the instance.
- Service:  `Process Integration Runtime`
- Plan: `api`
- Instance Name: `flashpipe-instance`

Click `Next`.
![InstanceDetails](images/oauth-client/cf/03_instance_details.png)

### 4. Enter required roles
Leave the default grant type to `client_credentials`. Select the roles shown below using the dropdown menu.

Click `Next`.
![InstanceRoles](images/oauth-client/cf/04_instance_roles.png)

### 5. Review and create instance
Review the details and click `Create`.
![Review](images/oauth-client/cf/05_instance_create.png)

### 6. Wait for creation to complete
![WIP](images/oauth-client/cf/06_instance_wip.png)

### 7. Create service key for instance
Once the instance has been create, click `***` its line and select `Create Service Key`. 
![CreateKey](images/oauth-client/cf/07_create_key.png)

### 8. Enter name of service key
Enter `flashpipe-key` as the name of the key.
![KeyDetails](images/oauth-client/cf/08_key_details.png)

### 9. View credentials of service key
Click on the created service key to view the credentials. Copy the following fields that will be needed for configuration with _FlashPipe_.
- `clientid`
- `clientsecret`
- `tokenurl`
![OAuthDetails](images/oauth-client/cf/09_oauth_details.png)