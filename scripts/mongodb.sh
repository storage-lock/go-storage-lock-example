#!/usr/bin/env bash
# 起一个本地 MongoDB 用于示例与测试。
#
# 同 mariadb.sh：凭据只从环境变量来，默认值仅指向本地临时容器。
set -euo pipefail

MONGO_ROOT_PASSWORD="${MONGO_ROOT_PASSWORD:-example-only-local}"
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-27017}"

docker run -d -p "${DB_PORT}:27017" --name storage-lock-mongodb \
  -e MONGO_INITDB_ROOT_USERNAME=root \
  -e MONGO_INITDB_ROOT_PASSWORD="${MONGO_ROOT_PASSWORD}" mongo

export STORAGE_LOCK_MONGO_URI="mongodb://root:${MONGO_ROOT_PASSWORD}@${DB_HOST}:${DB_PORT}/?connectTimeoutMS=300000"
echo "STORAGE_LOCK_MONGO_URI=$STORAGE_LOCK_MONGO_URI"
