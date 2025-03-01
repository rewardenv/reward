#!/bin/bash
# Common functions for scripts

# Prevent direct execution of this file
if [[ "${BASH_SOURCE[0]}" -ef "$0" ]]; then
  echo "This script should be sourced, not executed directly"
  exit 1
fi

# Guard against double-sourcing
if [[ -n "${_FUNCTIONS_SOURCED:-}" ]]; then
  return 0
fi
readonly _FUNCTIONS_SOURCED=1

# Default values
START_TIME=$(date +%s)
readonly START_TIME
: "${RETRY_INTERVAL:=1}"
: "${TIMEOUT:=600}"

# Print log messages
log() {
  [[ "${SILENT:-false}" != "true" ]] && printf "%s INFO: %s\n" "$(date --iso-8601=seconds)" "$*"
}

# Print error messages and exit
error() {
  exitcode=$?
  # color red
  printf "\033[1;31m%s ERROR: %s\033[0m\n" "$(date --iso-8601=seconds)" "$*" >&2

  stacktrace 2

  if [[ "${exitcode}" -eq 0 ]]; then exit 1; fi
  exit "${exitcode}"
}

stacktrace() {
  local i=1 line file func
  if [[ ${1:-0} -gt 0 ]]; then
    i=$1
  fi

  counter=1
  while read -r line func file < <(caller "$i"); do
    echo >&2 "[${counter}] ${file}:${line} ${func}(): $(sed -n "${line}p" "${file}")"
    ((counter++))
    ((i++))
  done
}

# Print error messages and exit with the command that failed
trapinfo() {
  # shellcheck disable=SC2145
  error "A command has failed. Exiting the script. COMMAND=($BASH_COMMAND) STATUS=($?)"
}

lock_acquire() {
  local lockfile="${1:-}"

  if [[ -z "$lockfile" ]]; then
    error "Lock file not provided"
  fi

  # Check if lockfile exists and process is still running
  if [[ -f "$lockfile" ]]; then
    local pid
    pid=$(cat "$lockfile" 2>/dev/null)
    if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
      error "Another process is already running, exiting"
    fi

    # Lock file exists but process is dead, remove stale lock
    log "Removing stale lock file"
    rm -f "$lockfile"
  fi

  # Check if lockfile basedir exists
  if [[ ! -d "$(dirname "$lockfile")" ]]; then
    mkdir -p "$(dirname "$lockfile")"
  fi

  # Create lock file with current PID
  echo $$ >"$lockfile"
}

lock_release() {
  local lockfile="${1:-}"

  if [[ -z "$lockfile" ]]; then
    error "Lock file not provided"
  fi

  # Only remove if we own the lock
  if [[ -f "$lockfile" ]] && [[ "$(cat "$lockfile" 2>/dev/null)" == "$$" ]]; then
    rm -f "$lockfile"
  fi
}

lock_cleanup() {
  local lockfile="${1:-}"

  if [[ -z "$lockfile" ]]; then
    error "Lock file not provided"
  fi

  lock_release "$lockfile"
}

conditional_sleep() {
  if [[ "${SLEEP:-false}" == "true" ]]; then
    sleep infinity
  elif [[ "${SLEEP:-false}" =~ ^[0-9]+$ ]]; then
    sleep "${SLEEP}"
  fi
}

shared_config_path() {
  if [[ -d "${SHARED_CONFIG_PATH:-/config}" ]] && [[ -w "${SHARED_CONFIG_PATH:-/config}" ]]; then
    echo "${SHARED_CONFIG_PATH:-/config}"
  else
    echo "/tmp"
  fi
}

app_path() {
  echo "${APP_PATH:-/var/www/html}"
}

# Compare versions
version_gt() { test "$(printf "%s\n" "${@#v}" | sort -V | head -n 1)" != "${1#v}"; }

# Check if command exists
check_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    error "Error: $1 is required but not installed."
  fi
}

run_hooks() {
  local hook="${1:-}"
  if [[ -n "${hook}" ]] && [[ -d "$(app_path)/hooks/${hook}.d" ]]; then
    for file in "$(app_path)"/hooks/"${hook}.d"/*.sh; do
      log "Running ${file} for ${hook}"
      # shellcheck disable=SC1090
      source "${file}"
    done
  fi
}

publish_shared_files() {
  IFS=':' read -ra shared_files_array <<<"${_shared_files:-}"
  if [[ "${#shared_files_array[@]}" == 0 ]]; then
    return
  fi

  for file in "${shared_files_array[@]}"; do
    if [[ -f "$(app_path)/${file}" ]] && [[ ! -L "$(app_path)/${file}" ]]; then
      log "Publishing shared file: ${file}"

      mkdir -p "$(shared_config_path)/$(dirname "${file}")"
      rm -f "$(shared_config_path)/${file}"
      cp -a "$(app_path)/${file}" "$(shared_config_path)/${file}"
    fi
  done

  unset _shared_files
}

link_shared_files() {
  IFS=':' read -ra shared_files_array <<<"${_shared_files:-}"

  if [[ "${#shared_files_array[@]}" == 0 ]]; then
    return
  fi

  for file in "${shared_files_array[@]}"; do
    if [[ -f "$(shared_config_path)/${file}" ]]; then
      log "Linking shared file: ${file}"

      # Create the directory of the file
      mkdir -p "$(app_path)/$(dirname "${file}")"
      rm -f "$(app_path)/${file}"
      ln -sf "$(shared_config_path)/${file}" "$(app_path)/${file}"
    fi
  done

  unset _shared_files
}

# Check if a command exists
check_timeout() {
  local current_time
  current_time=$(date +%s)
  local elapsed=$((current_time - START_TIME))

  if [[ "$elapsed" -ge "$TIMEOUT" ]]; then
    error "Global timeout of ${TIMEOUT}s reached"
  fi
}

# Main check function with retry logic
check_dependency() {
  set +u

  local func_name="$1"
  local attempt=1

  # Skip if already successful
  if [[ ${check_status[$func_name]:-false} == true ]]; then
    log "Skipping $func_name - already successful"

    set -u
    return 0
  fi

  while check_timeout; do
    log "Checking $func_name (attempt $attempt)"

    if ${func_name}; then
      log "Check succeeded: ${func_name}"
      check_status[${func_name}]=true

      set -u
      return 0
    fi

    attempt=$((attempt + 1))
    log "Check failed: ${func_name}, retrying in ${RETRY_INTERVAL}s..."
    sleep "${RETRY_INTERVAL}"
  done

  # If we get here, we've timed out
  log "Global Timeout reached ${func_name}"

  set -u
  return 1
}
