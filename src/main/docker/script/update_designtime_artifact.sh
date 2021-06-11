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
    echo "[ERROR] Mandatory environment variable $env_var_name is not populated"
    exit 1
  fi
}

function diff_directories() {
  local directory=$1
  local tenant_dir=$2
  local git_dir=$3
  local diff_found
  echo "[INFO] Checking for changes in $directory directory"
  echo "[INFO] Executing command: - diff --strip-trailing-cr -qr $tenant_dir/download/$directory/ $git_dir/$directory/"
  diffoutput="$(diff --strip-trailing-cr -qr "$tenant_dir/download/$directory/" "$git_dir/$directory/")"
  if [ -z "$diffoutput" ]; then
    echo "[INFO] No changes found in $directory directory"
  else
    echo "[INFO] Changes found in $directory directory"
    diff --strip-trailing-cr -r "$tenant_dir/download/$directory/" "$git_dir/$directory/"
    diff_found=1
  fi
  return $diff_found
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
    echo "[ERROR] Execution of java command failed"
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
  working_dir="/tmp"
else
  working_dir=$WORK_DIR
fi
git_src_dir=$GIT_DIR

# ----------------------------------------------------------------
# Use specific MANIFEST.MF and/or parameters.prop file (typically when moving to different environment)
# ----------------------------------------------------------------
if [ -n "$PARAM_FILE" ]; then
  echo "[INFO] Using $PARAM_FILE as parameters.prop file"
  cp "$PARAM_FILE" "$git_src_dir/src/main/resources/parameters.prop" || exit 1
fi
if [ -n "$MANIFEST_FILE" ]; then
  echo "[INFO] Using $MANIFEST_FILE as MANIFEST.MF file"
  cp "$MANIFEST_FILE" "$git_src_dir/META-INF/MANIFEST.MF" || exit 1
fi

# ----------------------------------------------------------------
# Set classpath for Java execution
# ----------------------------------------------------------------
if [ -z "$CLASSPATH_DIR" ]; then
  source /usr/bin/set_classpath.sh
else
  echo "[INFO] Using $CLASSPATH_DIR as classpath base directory "
  echo "[INFO] Setting WORKING_CLASSPATH environment variable"
  #  FLASHPIPE_VERSION
  export WORKING_CLASSPATH=$CLASSPATH_DIR/repository/io/github/engswee/flashpipe/2.0.0/flashpipe-2.0.0.jar
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
  echo "[INFO] IFlow will be updated (where necessary)"

  # 1 - Download IFlow from tenant
  zip_file="$working_dir/$IFLOW_ID.zip"
  echo "[INFO] Download existing IFlow from tenant for comparison"
  export OUTPUT_FILE=$zip_file
  export IFLOW_VER=active
  exec_java_command io.github.engswee.flashpipe.cpi.exec.DownloadDesignTimeArtifact

  # 2 - Diff contents from tenant against Git
  iflow_src_diff_found=

  cd "$working_dir" || exit 1
  echo "[INFO] Working in directory $working_dir"
  rm -rf "$working_dir/download"

  echo "[INFO] Unzipping downloaded IFlow artifact"
  echo "[INFO] Executing command: - /usr/bin/unzip -d $working_dir/download $zip_file"
  /usr/bin/unzip -d "$working_dir/download" "$zip_file"

  # Diff src/main/resources directory
  echo "[INFO] Removing commented lines from src/main/resources/parameters.prop files before comparison"
  sed -i '/^#/d' "$working_dir/download/src/main/resources/parameters.prop"
  sed -i '/^#/d' "$git_src_dir/src/main/resources/parameters.prop"

  diff_directories src/main/resources "$working_dir" "$git_src_dir"
  iflow_src_diff_found=$?

  # 3 - If there are differences, then update the IFlow
  if [[ "$iflow_src_diff_found" == "1" ]]; then
    echo '[INFO] IFlow will be updated in CPI tenant'
    # Clean up previous uploads
    rm -rf "$working_dir/upload"
    mkdir "$working_dir/upload" "$working_dir/upload/src" "$working_dir/upload/src/main"
    cp -r "$git_src_dir"/META-INF "$working_dir/upload"
    cp -r "$git_src_dir"/src/main/resources "$working_dir/upload/src/main"
    tenant_iflow_version=$(awk '/Bundle-Version/ {print $2}' "$working_dir/download/META-INF/MANIFEST.MF")
    export IFLOW_DIR="$working_dir/upload"
    export CURR_IFLOW_VER=$tenant_iflow_version
    exec_java_command io.github.engswee.flashpipe.cpi.exec.UpdateDesignTimeArtifact
    echo '[INFO] IFlow updated successfully'
  fi

# ----------------------------------------------------------------
# (B) IFlow does not exist in tenant, so upload the version from Git
# ----------------------------------------------------------------
elif [[ "$check_iflow_status" == "99" ]]; then
  echo "[INFO] IFlow will be uploaded to tenant"
  # Clean up previous uploads
  rm -rf "$working_dir/upload"
  mkdir "$working_dir/upload" "$working_dir/upload/src" "$working_dir/upload/src/main"
  cp -r "$git_src_dir"/META-INF "$working_dir/upload"
  cp -r "$git_src_dir"/src/main/resources "$working_dir/upload/src/main"
  export IFLOW_DIR="$working_dir/upload"
  exec_java_command io.github.engswee.flashpipe.cpi.exec.UploadDesignTimeArtifact
  echo '[INFO] IFlow created successfully'
fi