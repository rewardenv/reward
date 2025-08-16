#!/bin/bash
set -e

: "${VCL_TEMPLATE:=default}"

# Generate varnish config
gomplate <"/etc/varnish/${VCL_TEMPLATE}.vcl.template" >/etc/varnish/default.vcl

# If the first arg is `-D` or `--some-option` pass it to supervisord.
if [[ $# -eq 0 ]] || [[ "${1#-}" != "$1" ]] || [[ "${1#-}" != "$1" ]]; then
  set -- supervisord -c /etc/supervisor/supervisord.conf "$@"
# If the first arg is supervisord call it normally.
elif [[ "${1}" == "supervisord" ]]; then
  set -- "$@"
# If the first arg is anything else
else
  set -- "$@"
fi

exec "$@"
