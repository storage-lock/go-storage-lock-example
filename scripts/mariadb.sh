#!/usr/bin/env bash
# 起一个本地 MariaDB 用于示例与测试。
#
# 凭据一律从环境变量读取，不写死在仓库里：
# 示例仓一旦公开，任何写进文件的口令都等于永久泄露（git 历史删不干净）。
# 默认值只用于本地 127.0.0.1 的临时容器，不指向任何真实部署。
set -euo pipefail

MARIADB_ROOT_PASSWORD="${MARIADB_ROOT_PASSWORD:-example-only-local}"
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-3306}"

docker run -p "${DB_PORT}:3306" --name storage-lock-mariadb \
  -e MARIADB_ROOT_PASSWORD="${MARIADB_ROOT_PASSWORD}" -d mariadb:latest

export STORAGE_LOCK_MARIA_DSN="root:${MARIADB_ROOT_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/storage_lock_test"
echo "STORAGE_LOCK_MARIA_DSN=$STORAGE_LOCK_MARIA_DSN"
