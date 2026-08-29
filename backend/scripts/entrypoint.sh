#!/bin/sh
set -e

migrate -path ./migrations \
  -database "mysql://${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}" \
  up

exec "$@"
