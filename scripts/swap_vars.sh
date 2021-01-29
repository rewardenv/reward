#!/usr/bin/env bash
#===============================================================================
#          FILE:  swap_vars.sh
#
#         USAGE:  ./swap_vars.sh
#
#        AUTHOR:  mixe3y (Janos Miko), janos.miko@itg.cloud
#       COMPANY:  ITG
#       VERSION:  1.1
#       CREATED:  01/12/2021 12:00:00 CET
#===============================================================================
[ "$DEBUG" == "true" ] && set -x
set -eo pipefail

APP_NAME="REWARD"
PREV_APP_NAME="warden"
APP_NAME_LC=$(echo "$APP_NAME" | awk '{print tolower($0)}')
CREATE_BACKUP="false"
IMAGE_BASE="image: docker.io/wardenenv"
IMAGE_BASE_NEW='image: {{default "docker.io/rewardenv" .reward_docker_image_base}}'

if [ "$#" -ne 0 ]; then
  echo "We don't expect parameters..."
  exit 1
fi

readonly BASE_DIR="$(
  cd "$(
    dirname "$(
      (readlink "${BASH_SOURCE[0]}" || echo "${BASH_SOURCE[0]}") |
        sed -e "s#^../#$(dirname "$(dirname "${BASH_SOURCE[0]}")")/#"
    )"
  )" >/dev/null &&
    pwd
)/.."
pushd "${BASE_DIR}" >/dev/null

find "./templates" -type f -iname "*.yml" -print0 | while IFS= read -r -d '' FILE; do
  CMD_OUT=$(
    # shellcheck disable=SC2196
    # shellcheck disable=SC2016
    { egrep --color=no -o '\${1}\{[^\}]+\}' "$FILE" || test $? = 1; } | sort | uniq |
      awk -v APP_NAME=$APP_NAME '
        BEGIN {
          FS=":-";
          print "declare -a VARS_TO_SWAP; export VARS_TO_SWAP=("
        };
        {
          printf "'\''%s={{", $0;
          if (NF == 1)
          {
            tmp = $1
            gsub(/WARDEN/, APP_NAME, tmp);
            gsub(/\$\{/, "", tmp);
            sub("}", "", tmp);
            printf ".%s}}'\''\n", tolower(tmp);
          }
          if (NF >= 2)
          {
            tmp1 = $1
            tmp2 = $2
            gsub(/WARDEN/, APP_NAME, tmp1);
            gsub(/\$\{/,"", tmp1);
            gsub("}","",tmp2);
            printf "default \"%s\" .%s}}'\''\n", tmp2, tolower(tmp1);
          }
        };
        END {
          print ")";
        };'
  )

  eval "${CMD_OUT[@]}"

  [ "$CREATE_BACKUP" == "true" ] && cp -a "${FILE}" "${FILE}.old"

  for VAR in "${VARS_TO_SWAP[@]}"; do
    KEY="${VAR%%=*}"
    VALUE="${VAR##*=}"

    # don't swap variable if its $${VAR}
    sed -i.tmp -e "s@\([^\$]\)$KEY@\1$VALUE@g" "$FILE"
  done

  unset VARS_TO_SWAP
  rm -f "${FILE}.tmp"

  sed -i.tmp -e "/image/!s@${PREV_APP_NAME}@${APP_NAME_LC}@g" "$FILE"
  sed -i.tmp -e "s@${IMAGE_BASE}@${IMAGE_BASE_NEW}@g" "$FILE"
  rm -f "${FILE}.tmp"
done

# Global SVC specific changes
FILE="./templates/_services/docker-compose.yml"

if [ -f $FILE ]; then
  sed -i.tmp -e "s@WARDEN@${APP_NAME}@g" "$FILE"
  sed -i.tmp -e "s@Warden@${APP_NAME}@g" "$FILE"
  rm -f "${FILE}.tmp"
fi

# Global docker-compose.yml specific changes
FILE="./templates/_services/docker-compose.yml"

if [ -f $FILE ]; then
  sed -i.tmp -e "s@{{.reward_home_dir}}@.@g" "$FILE"
  rm -f "${FILE}.tmp"
fi

# Create templates for windows based on darwin
find "./templates" -type f -iname "*darwin*" -print0 | while IFS= read -r -d '' FILE; do
  NEWFILE="${FILE//darwin/windows}"

  # Check if file already exist
  if [ ! -f "$NEWFILE" ]; then
    cp -va "${FILE}" "${NEWFILE}"
  fi
done
