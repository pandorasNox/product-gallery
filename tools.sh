#!/usr/bin/env bash

set -o errexit
set -o nounset
# set -o xtrace

if set +o | grep -F 'set +o pipefail' > /dev/null; then
  # shellcheck disable=SC3040
  set -o pipefail
fi

if set +o | grep -F 'set +o posix' > /dev/null; then
  # shellcheck disable=SC3040
  set -o posix
fi

# -----------------------------------------------------------------------------

#   up                ...
#   down              ...
__usage="
Usage: $(basename $0) [OPTIONS]

Options:
  down              ...
  migrate-local     run migration script locally (via docker exec)
"

# -----------------------------------------------------------------------------

# function func_up() {
#   d8s up
#   d8s run tilt up
# }

# function func_down() {
#   d8s run tilt down
#   d8s down
# }

function func_migrate_local() {
  docker compose exec migrate /scripts/migrate.sh
}

function func_gen_css() {
  docker compose exec cli-css-generator npx tailwindcss -i /templates/css/input.css -o /static/assets/css/generated/output.css
}

# -----------------------------------------------------------------------------

if [ -z "$*" ]
then
  echo "$__usage"
else
    if [ $1 == "--help" ] || [ $1 == "-h" ]
    then
        echo "$__usage"
    fi

    # if [ $1 == "up" ]
    # then
    #   func_up
    # fi

    # if [ $1 == "down" ]
    # then
    #   func_down
    # fi

    if [ $1 == "migrate-local" ]
    then
      func_migrate_local 
    fi

    if [ $1 == "gen-css" ]
    then
      func_gen_css 
    fi

fi
