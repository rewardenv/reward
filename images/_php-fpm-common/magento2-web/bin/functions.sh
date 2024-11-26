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
