#!/bin/bash

# Arguments are passed to the script (and subsequently the Java commands)
# via environment variables. Below are the list of the variables:
#
# 1. Tenant details and credentials:
# HOST_TMN - Base URL for tenant management node of Cloud Integration (excluding the https:// prefix)
# BASIC_USERID - User ID (required when using Basic Authentication)
# BASIC_PASSWORD - Password (required when using Basic Authentication)
# HOST_OAUTH - Host name for OAuth authentication server (required when using OAuth Authentication)
# HOST_OAUTH_PATH - [Optional] Specific path for OAuth token server if it differs from /oauth/token, for example /oauth2/api/v1/token for Neo environments
# OAUTH_CLIENTID - OAuth Client ID (required when using OAuth Authentication)
# OAUTH_CLIENTSECRET - OAuth Client Secret (required when using OAuth Authentication)
#
# 2. Mandatory variables:
# IFLOW_ID - Comma separated list of Integration Flow IDs
#
# 3. Optional variables:
# DELAY_LENGTH - Delay (in seconds) between each check of IFlow deployment status (default to 30 if not set)
# MAX_CHECK_LIMIT - Max number of times to check for IFlow deployment status (default to 10 if not set)

function check_mandatory_env_var() {
  local env_var_name=$1
  local env_var_value=$2
  if [ -z "$env_var_value" ]; then
    echo "[ERROR] Mandatory environment variable $env_var_name is not populated"
    exit 1
  fi
}

check_mandatory_env_var "HOST_TMN" "$HOST_TMN"
if [ -z "$HOST_OAUTH" ]; then
  # Basic Auth
  check_mandatory_env_var "BASIC_USERID" "$BASIC_USERID"
  check_mandatory_env_var "BASIC_PASSWORD" "$BASIC_PASSWORD"
else
  # OAuth
  check_mandatory_env_var "OAUTH_CLIENTID" "$OAUTH_CLIENTID"
  check_mandatory_env_var "OAUTH_CLIENTSECRET" "$OAUTH_CLIENTSECRET"
fi
check_mandatory_env_var "IFLOW_ID" "$IFLOW_ID"

# Set debug log4j config
if [[ "$DEBUG" == "FLASHPIPE" ]]; then
  LOG4J_FILE='/tmp/log4j2-config/log4j2-debug-flashpipe.xml'
elif [[ "$DEBUG" == "APACHE" ]]; then
  LOG4J_FILE='/tmp/log4j2-config/log4j2-debug-apache.xml'
elif [[ "$DEBUG" == "ALL" ]]; then
  LOG4J_FILE='/tmp/log4j2-config/log4j2-debug-all.xml'
fi

source /usr/bin/set_classpath.sh

echo "[INFO] Deploying design time IFlow $IFLOW_ID to tenant runtime"
if [ -z "$LOG4J_FILE" ]; then
  echo "[INFO] Executing command: java -classpath $WORKING_CLASSPATH io.github.engswee.flashpipe.cpi.exec.DeployDesignTimeArtifact"
  java -classpath "$WORKING_CLASSPATH" io.github.engswee.flashpipe.cpi.exec.DeployDesignTimeArtifact
else
  echo "[INFO] Executing command: java -Dlog4j.configurationFile=$LOG4J_FILE -classpath $WORKING_CLASSPATH io.github.engswee.flashpipe.cpi.exec.DeployDesignTimeArtifact"
  java -Dlog4j.configurationFile="$LOG4J_FILE" -classpath "$WORKING_CLASSPATH" io.github.engswee.flashpipe.cpi.exec.DeployDesignTimeArtifact
fi