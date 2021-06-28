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
# IFLOW_NAME - Name of Integration Flow
# PACKAGE_ID - ID of Integration Package
# PACKAGE_NAME - Name of Integration Package
# GIT_DIR - directory containing contents of Integration Flow
#
# 3. Optional variables:
# PARAM_FILE - Use to a different parameters.prop file instead of the default in src/main/resources/
# MANIFEST_FILE - Use to a different MANIFEST.MF file instead of the default in META-INF/
# WORK_DIR - Working directory for in-transit files (default is /tmp if not set)

function check_mandatory_env_var() {
  local env_var_name=$1
  local env_var_value=$2
  if [ -z "$env_var_value" ]; then
    echo "[ERROR] 🛑 Mandatory environment variable $env_var_name is not populated"
    exit 1
  fi
}

function exec_java_command() {
  local return_code
  if [ -z "$LOG4J_FILE" ]; then
    echo "[INFO] Executing command: java -classpath $WORKING_CLASSPATH" "$@"
    java -classpath "$WORKING_CLASSPATH" "$@"
  else
    echo "[INFO] Executing command: java -Dlog4j.configurationFile=$LOG4J_FILE -classpath $WORKING_CLASSPATH" "$@"
    java -Dlog4j.configurationFile="$LOG4J_FILE" -classpath "$WORKING_CLASSPATH" "$@"
  fi
  return_code=$?
  if [[ "$return_code" == "1" ]]; then
    echo "[ERROR] 🛑 Execution of java command failed"
    exit 1
  fi
  return $return_code
}

# ----------------------------------------------------------------
# Check presence of environment variables
# ----------------------------------------------------------------
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
check_mandatory_env_var "GIT_DIR" "$GIT_DIR"
check_mandatory_env_var "IFLOW_ID" "$IFLOW_ID"
check_mandatory_env_var "IFLOW_NAME" "$IFLOW_NAME"
check_mandatory_env_var "PACKAGE_ID" "$PACKAGE_ID"
check_mandatory_env_var "PACKAGE_NAME" "PACKAGE_NAME"

if [ -z "$WORK_DIR" ]; then
  WORK_DIR="/tmp"
fi

# ----------------------------------------------------------------
# Use specific MANIFEST.MF and/or parameters.prop file (typically when moving to different environment)
# ----------------------------------------------------------------
if [ -n "$PARAM_FILE" ]; then
  echo "[INFO] Using $PARAM_FILE as parameters.prop file"
  cp "$PARAM_FILE" "$GIT_DIR/src/main/resources/parameters.prop" || exit 1
fi
if [ -n "$MANIFEST_FILE" ]; then
  echo "[INFO] Using $MANIFEST_FILE as MANIFEST.MF file"
  cp "$MANIFEST_FILE" "$GIT_DIR/META-INF/MANIFEST.MF" || exit 1
fi

# ----------------------------------------------------------------
# Set classpath for Java execution
# ----------------------------------------------------------------
if [ -z "$CLASSPATH_DIR" ]; then
  source /usr/bin/set_classpath.sh
else
  echo "[INFO] Using $CLASSPATH_DIR as classpath base directory "
  echo "[INFO] Setting WORKING_CLASSPATH environment variable"
  FLASHPIPE_VERSION=2.1.0
  export WORKING_CLASSPATH=$CLASSPATH_DIR/repository/io/github/engswee/flashpipe/$FLASHPIPE_VERSION/flashpipe-$FLASHPIPE_VERSION.jar
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

# ----------------------------------------------------------------
# Query for existence of IFlow
# ----------------------------------------------------------------
exec_java_command io.github.engswee.flashpipe.cpi.exec.QueryDesignTimeArtifact
check_iflow_status=$?

# ----------------------------------------------------------------
# (A) IFlow already exists in tenant, so check if it needs to be updated
# ----------------------------------------------------------------
if [[ "$check_iflow_status" == "0" ]]; then
  echo "[INFO] Checking if IFlow design needs to be updated"

  # 1 - Download IFlow from tenant
  zip_file="$WORK_DIR/$IFLOW_ID.zip"
  echo "[INFO] Download existing IFlow from tenant for comparison"
  export OUTPUT_FILE=$zip_file
  export IFLOW_VER=active
  exec_java_command io.github.engswee.flashpipe.cpi.exec.DownloadDesignTimeArtifact

  # 2 - Diff contents from tenant against Git
  cd "$WORK_DIR" || exit 1
  echo "[INFO] Working in directory $WORK_DIR"
  rm -rf "$WORK_DIR/download"

  echo "[INFO] Unzipping downloaded IFlow artifact"
  echo "[INFO] Executing command: - /usr/bin/unzip -d $WORK_DIR/download $zip_file"
  /usr/bin/unzip -d "$WORK_DIR/download" "$zip_file"

  # Any configured value will remain in IFlow even if the IFlow is replaced and the parameter is no longer used
  # Therefore diff of parameters.prop may come up with false differences
  echo "[INFO] Checking for changes in src/main/resources directory"
  echo "[INFO] Executing command: - diff --strip-trailing-cr -qr $WORK_DIR/download/src/main/resources/ $GIT_DIR/src/main/resources/"
  diffoutput="$(diff --strip-trailing-cr -qr -x 'parameters.prop' "$WORK_DIR/download/src/main/resources/" "$GIT_DIR/src/main/resources/")"
  if [ -z "$diffoutput" ]; then
    echo '[INFO] No changes detected. IFlow design does not need to be updated'
  else
    echo "[INFO] Changes found in src/main/resources directory"
    diff --strip-trailing-cr -r -x 'parameters.prop' "$WORK_DIR/download/src/main/resources/" "$GIT_DIR/src/main/resources/"
    echo '[INFO] IFlow design will be updated in CPI tenant'
    # Clean up previous uploads
    rm -rf "$WORK_DIR/upload"
    mkdir -p "$WORK_DIR/upload/src/main"
    cp -r "$GIT_DIR/META-INF" "$WORK_DIR/upload"
    cp -r "$GIT_DIR/src/main/resources" "$WORK_DIR/upload/src/main"
    tenant_iflow_version=$(awk '/Bundle-Version/ {print $2}' "$WORK_DIR/download/META-INF/MANIFEST.MF")
    export IFLOW_DIR="$WORK_DIR/upload"
    export CURR_IFLOW_VER=$tenant_iflow_version
    exec_java_command io.github.engswee.flashpipe.cpi.exec.UpdateDesignTimeArtifact
    echo '[INFO] 🏆 IFlow design updated successfully'
  fi

  # 4 - Update the configuration of the IFlow based on parameters.prop file
  echo '[INFO] Updating configured parameter(s) of IFlow where necessary'
  if [ -z "$PARAM_FILE" ]; then
    export PARAM_FILE="$GIT_DIR/src/main/resources/parameters.prop"
  fi
  exec_java_command io.github.engswee.flashpipe.cpi.exec.UpdateConfiguration

# ----------------------------------------------------------------
# (B) IFlow does not exist in tenant, so upload the version from Git
# ----------------------------------------------------------------
elif [[ "$check_iflow_status" == "99" ]]; then
  echo "[INFO] IFlow will be uploaded to tenant" # Clean up previous uploads
  rm -rf "$WORK_DIR/upload"
  mkdir -p "$WORK_DIR/upload/src/main"
  cp -r "$GIT_DIR/META-INF" "$WORK_DIR/upload"
  cp -r "$GIT_DIR/src/main/resources" "$WORK_DIR/upload/src/main"
  export IFLOW_DIR="$WORK_DIR/upload"
  exec_java_command io.github.engswee.flashpipe.cpi.exec.UploadDesignTimeArtifact
  echo '[INFO] 🏆 IFlow created successfully'
fi