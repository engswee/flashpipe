#!/bin/bash

# Arguments are passed to the script (and subsequently the Java commands)
# via environment variables. Below are the list of the variables:
#
# 1. Tenant details and credentials:
# HOST_TMN - Base URL for tenant management node of Cloud Integration (excluding the https:// prefix)
# BASIC_USERID - User ID (required when using Basic Authentication)
# BASIC_PASSWORD - Password (required when using Basic Authentication)
# HOST_OAUTH - Host name for OAuth authentication server (required when using OAuth Authentication)
# OAUTH_CLIENTID - OAuth Client ID (required when using OAuth Authentication)
# OAUTH_CLIENTSECRET - OAuth Client Secret (required when using OAuth Authentication)
#
# 2. Mandatory variables:
# IFLOW_ID - ID of Integration Flow
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

if [ -z "$CLASSPATH_DIR" ]; then
  source /usr/bin/set_classpath.sh
else
  echo "[INFO] Using $CLASSPATH_DIR as classpath base directory "
  echo "[INFO] Setting WORKING_CLASSPATH environment variable"
  #  FLASHPIPE_VERSION
  export WORKING_CLASSPATH=$CLASSPATH_DIR/repository/io/github/engswee/flashpipe/2.0.1/flashpipe-2.0.1.jar
  export WORKING_CLASSPATH=$WORKING_CLASSPATH:$CLASSPATH_DIR/repository/org/codehaus/groovy/groovy-all/2.4.12/groovy-all-2.4.12.jar
  export WORKING_CLASSPATH=$WORKING_CLASSPATH:$CLASSPATH_DIR/repository/org/apache/httpcomponents/core5/httpcore5/5.0.4/httpcore5-5.0.4.jar
  export WORKING_CLASSPATH=$WORKING_CLASSPATH:$CLASSPATH_DIR/repository/org/apache/httpcomponents/client5/httpclient5/5.0.4/httpclient5-5.0.4.jar
  export WORKING_CLASSPATH=$WORKING_CLASSPATH:$CLASSPATH_DIR/repository/commons-codec/commons-codec/1.15/commons-codec-1.15.jar
  export WORKING_CLASSPATH=$WORKING_CLASSPATH:$CLASSPATH_DIR/repository/org/slf4j/slf4j-api/1.7.25/slf4j-api-1.7.25.jar
  export WORKING_CLASSPATH=$WORKING_CLASSPATH:$CLASSPATH_DIR/repository/org/apache/logging/log4j/log4j-slf4j-impl/2.14.1/log4j-slf4j-impl-2.14.1.jar
  export WORKING_CLASSPATH=$WORKING_CLASSPATH:$CLASSPATH_DIR/repository/org/apache/logging/log4j/log4j-api/2.14.1/log4j-api-2.14.1.jar
  export WORKING_CLASSPATH=$WORKING_CLASSPATH:$CLASSPATH_DIR/repository/org/apache/logging/log4j/log4j-core/2.14.1/log4j-core-2.14.1.jar
  export WORKING_CLASSPATH=$WORKING_CLASSPATH:$CLASSPATH_DIR/repository/org/zeroturnaround/zt-zip/1.14/zt-zip-1.14.jar
fi

echo "[INFO] Deploying design time IFlow $IFLOW_ID to tenant runtime"
if [ -z "$LOG4J_FILE" ]; then
  echo "[INFO] Executing command: java -classpath $WORKING_CLASSPATH io.github.engswee.flashpipe.cpi.exec.DeployDesignTimeArtifact"
  java -classpath "$WORKING_CLASSPATH" io.github.engswee.flashpipe.cpi.exec.DeployDesignTimeArtifact
else
  echo "[INFO] Executing command: java -Dlog4j.configurationFile=$LOG4J_FILE -classpath $WORKING_CLASSPATH io.github.engswee.flashpipe.cpi.exec.DeployDesignTimeArtifact"
  java -Dlog4j.configurationFile="$LOG4J_FILE" -classpath "$WORKING_CLASSPATH" io.github.engswee.flashpipe.cpi.exec.DeployDesignTimeArtifact
fi