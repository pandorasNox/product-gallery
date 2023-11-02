#!/usr/bin/env sh

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

timeout 120 sh -c "while ! timeout 1 ash -c 'nc -z ${POSTGRES_HOST} ${POSTGRES_PORT}'; do sleep 1; printf '%s' '.'; done";
echo √ postgres port open;

#migrate -h
export POSTGRESQL_URL="postgres://${POSTGRES_USER}:${PGPASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable"

set -x

migrate -database ${POSTGRESQL_URL} -path /migrations up

echo done;
