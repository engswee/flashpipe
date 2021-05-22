#!/bin/bash

function die() {
  printf '%s\n' "$1" >&2
  exit 1
}

function usage() {
  echo -e "usage: update_designtime_artifact.sh [--logcfgfile=<path_to_file>] [--param_file=<path_to_file>] [--manifest_file=<path_to_file>] [--debug] [--classpath_base_dir=<path_to_dir>] working_dir tmn_host cpi_user cpi_password iflow_id iflow_name package_id package_name git_src_dir\n"
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
    if [ -z "$debug" ]; then
      echo "$diffoutput"
    else
      diff --strip-trailing-cr -r "$tenant_dir/download/$directory/" "$git_dir/$directory/"
    fi
    diff_found=1
  fi
  return $diff_found
}

function exec_java_command() {
  local return_code
  if [ -z "$logcfgfile" ]; then
    echo "[INFO] Executing command: java -classpath $WORKING_CLASSPATH" "$@"
    java -classpath "$WORKING_CLASSPATH" "$@"
  else
    echo "[INFO] Executing command: java -Dlog4j.configurationFile=$logcfgfile -classpath $WORKING_CLASSPATH" "$@"
    java -Dlog4j.configurationFile="$logcfgfile" -classpath "$WORKING_CLASSPATH" "$@"
  fi
  return_code=$?
  if [[ "$return_code" == "1" ]]; then
    echo "[ERROR] Execution of java command failed"
    exit 1
  fi
  return $return_code
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
  --param_file=?*)
    param_file=${1#*=} # Delete everything up to "=" and assign the remainder.
    ;;
  --param_file=) # Handle the case of an empty --param_file=
    die 'ERROR: "--param_file" requires a non-empty option argument.'
    ;;
  --manifest_file=?*)
    manifest_file=${1#*=} # Delete everything up to "=" and assign the remainder.
    ;;
  --manifest_file=) # Handle the case of an empty --manifest_file=
    die 'ERROR: "--manifest_file" requires a non-empty option argument.'
    ;;
  --debug)
    debug="X"
    ;;
  --classpath_base_dir=?*)
    classpath_base_dir=${1#*=} # Delete everything up to "=" and assign the remainder.
    ;;
  --classpath_base_dir=) # Handle the case of an empty --classpath_base_dir=
    die 'ERROR: "--classpath_base_dir" requires a non-empty option argument.'
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
#1 - Working directory
#2 - CPI tenant management host
#3 - CPI username
#4 - CPI password
#5 - IFlow ID
#6 - IFlow Name
#7 - Package ID
#8 - Directory for Git source
working_dir=$1
tmn_host=$2
cpi_user=$3
cpi_password=$4
iflow_id=$5
iflow_name=$6
package_id=$7
package_name=$8
git_src_dir=$9

if [ -z "$classpath_base_dir" ]; then
  source /usr/bin/set_classpath.sh
else
  echo "[INFO] Using $classpath_base_dir as classpath base directory "
  echo "[INFO] Setting WORKING_CLASSPATH environment variable"
  #  FLASHPIPE_VERSION
  export WORKING_CLASSPATH=$classpath_base_dir/repository/io/github/engswee/flashpipe/1.0.1/flashpipe-1.0.1.jar
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
exec_java_command io.github.engswee.flashpipe.cpi.exec.QueryDesignTimeArtifact "$iflow_id" "$package_id" "$tmn_host" "$cpi_user" "$cpi_password"
check_iflow_status=$?

# Use specific MANIFEST.MF and/or parameters.prop file (typically when moving to different environment)
if [ -n "$param_file" ]; then
  echo "[INFO] Using $param_file as parameters.prop file"
  cp "$param_file" "$git_src_dir/src/main/resources/parameters.prop"
fi
if [ -n "$manifest_file" ]; then
  echo "[INFO] Using $manifest_file as MANIFEST.MF file"
  cp "$manifest_file" "$git_src_dir/META-INF/MANIFEST.MF"
fi

if [[ "$check_iflow_status" == "0" ]]; then
  # (A)  IFlow already exists in tenant, so check if it needs to be updated
  echo "[INFO] IFlow will be updated (where necessary)"

  # 1 - Download IFlow from tenant
  zip_file="$working_dir/$iflow_id.zip"
  echo "[INFO] Download existing IFlow from tenant for comparison"
  exec_java_command io.github.engswee.flashpipe.cpi.exec.DownloadDesignTimeArtifact "$iflow_id" active "$tmn_host" "$cpi_user" "$cpi_password" "$zip_file"

  # 2 - Diff contents from tenant against Git
  iflow_src_diff_found=

  cd "$working_dir" || exit
  echo "[INFO] Working in directory $working_dir"
  rm -rf "$working_dir/download"

  echo "[INFO] Unzipping downloaded IFlow artifact"
  echo "[INFO] Executing command: - /usr/bin/unzip -d download $zip_file"
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
    exec_java_command io.github.engswee.flashpipe.cpi.exec.UpdateDesignTimeArtifact "$iflow_name" "$iflow_id" "$package_id" "$working_dir/upload" "$tenant_iflow_version" "$tmn_host" "$cpi_user" "$cpi_password"
    echo '[INFO] IFlow updated successfully'
  fi

elif [[ "$check_iflow_status" == "99" ]]; then
  #  (B) IFlow does not exist in tenant, so upload the version from Git
  echo "[INFO] IFlow will be uploaded to tenant"
  # Clean up previous uploads
  rm -rf "$working_dir/upload"
  mkdir "$working_dir/upload" "$working_dir/upload/src" "$working_dir/upload/src/main"
  cp -r "$git_src_dir"/META-INF "$working_dir/upload"
  cp -r "$git_src_dir"/src/main/resources "$working_dir/upload/src/main"
  exec_java_command io.github.engswee.flashpipe.cpi.exec.UploadDesignTimeArtifact "$iflow_name" "$iflow_id" "$package_id" "$package_name" "$working_dir/upload" "$tmn_host" "$cpi_user" "$cpi_password"
  echo '[INFO] IFlow created successfully'
fi
