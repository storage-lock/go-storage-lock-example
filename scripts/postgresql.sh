#!/usr/bin/env bash

# 启动测试使用的Postgresql数据库
docker run -d --name storage-lock-postgres -p 5432:5432 -e POSTGRES_PASSWORD=example-only-local postgres:14

export STORAGE_LOCK_POSTGRESQL_DSN="host=127.0.0.1 user=postgres password=example-only-local port=5432 dbname=postgres sslmode=disable"

