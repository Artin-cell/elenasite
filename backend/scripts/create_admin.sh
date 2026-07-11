#!/usr/bin/env bash

set -e

USERNAME="${1:?имя пользователя}"
PASSWORD="${2:?пароль}"

HASH=$(go run "$(dirname "$0")/bcrypt/main.go" "$PASSWORD")

psql "${DATABASE_URL:-postgres://postgres:root@localhost:5432/elena_db}" \
  -c "insert into admins (id, username, password_hash)
      values (uuid_generate_v4(), '$USERNAME', '$HASH')
      on conflict (username) do update set password_hash = excluded.password_hash;"

echo "Admin '$USERNAME' создан / обновлён."
