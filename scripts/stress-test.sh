#!/usr/bin/env sh
set -eu

BASE_URL="${1:-https://chain.fycbit.com}"
seq 1 120 | xargs -n1 -P60 -I{} sh -c "curl -s -o /dev/null -w '%{http_code}\n' '$BASE_URL/health'" | sort | uniq -c
