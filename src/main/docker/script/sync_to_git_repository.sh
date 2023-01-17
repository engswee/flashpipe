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
# PACKAGE_ID - ID of Integration Package
# GIT_SRC_DIR - Base directory containing contents of Integration Flow(s)
#
# 3. Optional variables:
# WORK_DIR - Working directory for in-transit files (default is /tmp if not set)
# DIR_NAMING_TYPE - Name IFlow directories by ID or Name
# DRAFT_HANDLING - Handling when IFlow is in draft version
# INCLUDE_IDS - List of included IFlow IDs
# EXCLUDE_IDS - List of excluded IFlow IDs
# COMMIT_MESSAGE - Message used in commit
# SCRIPT_COLLECTION_MAP - Comma-separated source-target ID pairs for converting script collection references during sync
# NORMALIZE_MANIFEST_ACTION - Action for normalizing IFlow ID & Name in MANIFEST.MF
# NORMALIZE_MANIFEST_PREFIX_SUFFIX - Prefix/suffix used for normalizing IFlow ID & Name in MANIFEST.MF
# SYNC_PACKAGE_LEVEL_DETAILS - Sync details of Integration Package
# NORMALIZE_PACKAGE_ACTION - Action for normalizing Package ID & Name package file
# NORMALIZE_PACKAGE_ID_PREFIX_SUFFIX - Prefix/suffix used for normalizing Package ID
# NORMALIZE_PACKAGE_NAME_PREFIX_SUFFIX - Prefix/suffix used for normalizing Package Name

function check_mandatory_env_var() {
  local env_var_name=$1
  local env_var_value=$2
  if [ -z "$env_var_value" ]; then
    echo "[ERROR] Mandatory environment variable $env_var_name is not populated"
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
check_mandatory_env_var "PACKAGE_ID" "$PACKAGE_ID"
check_mandatory_env_var "GIT_SRC_DIR" "$GIT_SRC_DIR"
if [ -z "$WORK_DIR" ]; then
  export WORK_DIR="/tmp"
fi

# Set debug log4j config
if [[ "$DEBUG" == "FLASHPIPE" ]]; then
  LOG4J_FILE='/tmp/log4j2-config/log4j2-debug-flashpipe.xml'
elif [[ "$DEBUG" == "APACHE" ]]; then
  LOG4J_FILE='/tmp/log4j2-config/log4j2-debug-apache.xml'
elif [[ "$DEBUG" == "ALL" ]]; then
  LOG4J_FILE='/tmp/log4j2-config/log4j2-debug-all.xml'
fi

if [ -z "$COMMIT_MESSAGE" ]; then
  export COMMIT_MESSAGE="Sync repo from tenant"
fi

source /usr/bin/set_classpath.sh

exec_java_command io.github.engswee.flashpipe.cpi.exec.DownloadIntegrationPackageContent

# Commit
echo "[INFO] Configuring git"
git config --global core.autocrlf input
git config --global user.email "41898282+github-actions[bot]@users.noreply.github.com"
git config --global user.name "github-actions[bot]"
git config --global --add safe.directory '*'
echo "[INFO] Adding all files for Git tracking"
git add --all --verbose
echo "[INFO] Trying to commit changes"
if git commit -m "$COMMIT_MESSAGE" -a --verbose; then
  echo "[INFO] 🏆 Changes committed"
else
  echo "[INFO] 🏆 No changes to commit"
fi
