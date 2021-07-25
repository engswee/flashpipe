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
# Parse arguments, overriding existing environment variables
# ----------------------------------------------------------------
echo "[INFO] parsing arguments"
for ARGUMENT in "$@"
do  
    KEY=$(echo $ARGUMENT | cut -f1 -d=)
    VALUE=$(echo $ARGUMENT | cut -f2 -d=)   

    case "$KEY" in
            HOST_TMN)        HOST_TMN=${VALUE} ;;
            BASIC_USERID)    BASIC_USERID=${VALUE} ;;     
            BASIC_PASSWORD)  BASIC_PASSWORD=${VALUE} ;;
            HOST_OAUTH)      HOST_OAUTH=${VALUE} ;;
            HOST_OAUTH_PATH) HOST_OAUTH_PATH=${VALUE} ;;
            OAUTH_CLIENTID)  OAUTH_CLIENTID=${VALUE} ;;
            OAUTH_CLIENTSECRET) OAUTH_CLIENTSECRET=${VALUE} ;;           
            LOG4J_FILE)      LOG4J_FILE=${VALUE} ;;
            PACKAGE_ID)      PACKAGE_ID=${VALUE} ;;
            GIT_SRC_DIR)     GIT_SRC_DIR=${VALUE} ;;
            DRAFT_HANDLING)  DRAFT_HANDLING=${VALUE} ;;
            INCLUDE_IDS)     INCLUDE_IDS=${VALUE} ;;
            EXCLUDE_IDS)     EXCLUDE_IDS=${VALUE} ;;
            WORK_DIR)        WORK_DIR=${VALUE} ;;
            DIR_NAMING_TYPE) DIR_NAMING_TYPE=${VALUE} ;;
            COMMIT_MESSAGE)  COMMIT_MESSAGE=${VALUE} ;;
            DO_NOT_COMMIT)   DO_NOT_COMMIT=${VALUE} ;;
            *)               echo "[INFO] Unknown argument: $ARGUMENT, ignoring"
    esac    
done

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
if [ -z "$DO_NOT_COMMIT" ]; then
  export DO_NOT_COMMIT=0
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

if [ -z "$CLASSPATH_DIR" ]; then
  source /usr/bin/set_classpath.sh
else
  echo "[INFO] Using $CLASSPATH_DIR as classpath base directory "
  echo "[INFO] Setting WORKING_CLASSPATH environment variable"
  FLASHPIPE_VERSION=2.2.2
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

# Git config
echo "[INFO] Configuring git"
git config --global core.autocrlf input
git config --local user.email "41898282+github-actions[bot]@users.noreply.github.com"
git config --local user.name "github-actions[bot]"

exec_java_command io.github.engswee.flashpipe.cpi.exec.DownloadIntegrationPackageContent

# Commit
if [[ "$DO_NOT_COMMIT" != "1" ]]; then
  echo "[INFO] Adding all files for Git tracking"
  git add --all --verbose
  echo "[INFO] Trying to commit changes"
  if git commit -m "$COMMIT_MESSAGE" -a --verbose; then
    echo "[INFO] 🏆 Changes committed"
  else
    echo "[INFO] 🏆 No changes to commit"
  fi
else
  echo "[INFO] 🏆 Changes not committed (DO_NOT_COMMIT=1)"
fi
