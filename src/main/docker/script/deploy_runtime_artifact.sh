#!/bin/bash

function die() {
  printf '%s\n' "$1" >&2
  exit 1
}

function usage() {
  echo -e "usage: deploy_runtime_artifact.sh [--logcfgfile=<pathtofile>] [--delay=<delay_in_sec>] [--maxcheck=<count>] [--classpath_base_dir=<path_to_dir>] iflow_id tmn_host cpi_user cpi_password\n"
}

logcfgfile=

while :; do
  case $1 in
  -h)
    usage
    exit
    ;;
  --logcfgfile=?*)
    logcfgfile=${1#*=} # Delete everything up to "=" and assign the remainder.
    ;;
  --logcfgfile=) # Handle the case of an empty --logcfgfile=
    die 'ERROR: "--logcfgfile" requires a non-empty option argument.'
    ;;
  --classpath_base_dir=?*)
    classpath_base_dir=${1#*=} # Delete everything up to "=" and assign the remainder.
    ;;
  --classpath_base_dir=) # Handle the case of an empty --classpath_base_dir=
    die 'ERROR: "--classpath_base_dir" requires a non-empty option argument.'
    ;;
  --delay=?*)
    delay=${1#*=} # Delete everything up to "=" and assign the remainder.
    ;;
  --delay=) # Handle the case of an empty --delay=
    die 'ERROR: "--delay" requires a non-empty option argument.'
    ;;
  --maxcheck=?*)
    maxcheck=${1#*=} # Delete everything up to "=" and assign the remainder.
    ;;
  --maxcheck=) # Handle the case of an empty --maxcheck=
    die 'ERROR: "--maxcheck" requires a non-empty option argument.'
    ;;
  --) # End of all options.
    shift
    break
    ;;
  *) # Default case: No more options, so break out of the loop.
    break ;;
  esac
  shift
done

#Arguments
#1 - IFlow ID
#2 - CPI tenant management host
#3 - CPI username
#4 - CPI password
iflow_id=$1
tmn_host=$2
cpi_user=$3
cpi_password=$4
if [ -z "$delay" ]; then
  delay_in=30
else
  delay_in=$delay
fi
if [ -z "$maxcheck" ]; then
  maxcheck_in=10
else
  maxcheck_in=$maxcheck
fi

if [ -z "$classpath_base_dir" ]; then
  source /usr/bin/set_classpath.sh
else
  echo "[INFO] Using $classpath_base_dir as classpath base directory "
  echo "[INFO] Setting WORKING_CLASSPATH environment variable"
  #  FLASHPIPE_VERSION
  export WORKING_CLASSPATH=$classpath_base_dir/repository/com/equalize/flashpipe/1.0.1/flashpipe-1.0.1.jar
  export WORKING_CLASSPATH=$WORKING_CLASSPATH:$classpath_base_dir/repository/org/codehaus/groovy/groovy-all/2.4.12/groovy-all-2.4.12.jar
  export WORKING_CLASSPATH=$WORKING_CLASSPATH:$classpath_base_dir/repository/org/apache/httpcomponents/core5/httpcore5/5.0.4/httpcore5-5.0.4.jar
  export WORKING_CLASSPATH=$WORKING_CLASSPATH:$classpath_base_dir/repository/org/apache/httpcomponents/client5/httpclient5/5.0.4/httpclient5-5.0.4.jar
  export WORKING_CLASSPATH=$WORKING_CLASSPATH:$classpath_base_dir/repository/commons-codec/commons-codec/1.15/commons-codec-1.15.jar
  export WORKING_CLASSPATH=$WORKING_CLASSPATH:$classpath_base_dir/repository/org/slf4j/slf4j-api/1.7.25/slf4j-api-1.7.25.jar
  export WORKING_CLASSPATH=$WORKING_CLASSPATH:$classpath_base_dir/repository/org/apache/logging/log4j/log4j-slf4j-impl/2.14.1/log4j-slf4j-impl-2.14.1.jar
  export WORKING_CLASSPATH=$WORKING_CLASSPATH:$classpath_base_dir/repository/org/apache/logging/log4j/log4j-api/2.14.1/log4j-api-2.14.1.jar
  export WORKING_CLASSPATH=$WORKING_CLASSPATH:$classpath_base_dir/repository/org/apache/logging/log4j/log4j-core/2.14.1/log4j-core-2.14.1.jar
  export WORKING_CLASSPATH=$WORKING_CLASSPATH:$classpath_base_dir/repository/org/zeroturnaround/zt-zip/1.14/zt-zip-1.14.jar
fi

echo "[INFO] Deploying design time IFlow $iflow_id to tenant runtime"
if [ -z "$logcfgfile" ]; then
  echo "[INFO] Executing command: java -classpath $WORKING_CLASSPATH com.equalize.flashpipe.cpi.exec.DeployDesignTimeArtifact $iflow_id $tmn_host $cpi_user $cpi_password $delay_in $maxcheck_in"
  java -classpath "$WORKING_CLASSPATH" com.equalize.flashpipe.cpi.exec.DeployDesignTimeArtifact "$iflow_id" "$tmn_host" "$cpi_user" "$cpi_password" $delay_in $maxcheck_in
else
  echo "[INFO] Executing command: java -Dlog4j.configurationFile=$logcfgfile -classpath $WORKING_CLASSPATH com.equalize.flashpipe.cpi.exec.DeployDesignTimeArtifact $iflow_id $tmn_host $cpi_user $cpi_password $delay_in $maxcheck_in"
  java -Dlog4j.configurationFile="$logcfgfile" -classpath "$WORKING_CLASSPATH" com.equalize.flashpipe.cpi.exec.DeployDesignTimeArtifact "$iflow_id" "$tmn_host" "$cpi_user" "$cpi_password" $delay_in $maxcheck_in
fi