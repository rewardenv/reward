#!/bin/bash
set -e

main() {
    # If the first arg is `-D` or `--some-option` pass it to php-fpm.
  if [[ "${1#-}" != "$1" ]]; then
    set -- php-fpm "$@"
  fi

  exec "$@"
}

(return 0 2>/dev/null) && sourced=1

if [[ -z "${sourced:-}" ]]; then
  main "$@"
fi
