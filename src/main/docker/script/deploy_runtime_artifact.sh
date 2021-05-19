#!/bin/bash

function die() {
  printf '%s\n' "$1" >&2
  exit 1
}

function usage() {
  echo -e "usage: deploy.sh [--logcfgfile=<pathtofile>] iflow_id tmn_host cpi_user cpi_password\n"
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

echo "[INFO] Deploying design time IFlow $iflow_id to tenant runtime"
source /usr/bin/set_classpath.sh
if [ -z "$logcfgfile" ]; then
  echo "[INFO] Executing command: java -classpath $WORKING_CLASSPATH com.equalize.flashpipe.cpi.exec.DeployDesignTimeArtifact $iflow_id $tmn_host $cpi_user $cpi_password"
  java -classpath $WORKING_CLASSPATH com.equalize.flashpipe.cpi.exec.DeployDesignTimeArtifact $iflow_id $tmn_host $cpi_user $cpi_password
else
  echo "[INFO] Executing command: java -Dlog4j.configurationFile=$logcfgfile -classpath $WORKING_CLASSPATH com.equalize.flashpipe.cpi.exec.DeployDesignTimeArtifact $iflow_id $tmn_host $cpi_user $cpi_password"
  java -Dlog4j.configurationFile=$logcfgfile -classpath $WORKING_CLASSPATH com.equalize.flashpipe.cpi.exec.DeployDesignTimeArtifact $iflow_id $tmn_host $cpi_user $cpi_password
fi