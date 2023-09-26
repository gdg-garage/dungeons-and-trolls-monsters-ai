#!/usr/bin/env bash

# Usage: go run main.go BOT_TOKEN | ./stdin2betterstack.sh BETTER_STACK_TOKEN
# check if the token is passed
if [ -z "$1" ]; then
  echo "Please provide Better Stack token."
  echo "USAGE: go run main.go BOT_TOKEN | ./stdin2betterstack.sh BETTER_STACK_TOKEN"
  exit 1
fi

batch=20
thisBatch=0
while IFS=$'\n' read -r line; do
  printf '%s\n' "$line"
  curl -X POST \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer '"$1" \
  -d "${line}" \
  -k \
  https://in.logs.betterstack.com
done
