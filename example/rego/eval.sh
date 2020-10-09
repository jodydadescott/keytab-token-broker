#!/bin/bash

nq="data.kbridge.auth_get_nonce"
kq="data.kbridge.auth_get_keytab"

function main() {

  failed=false

  query="$nq"
  run_test 1 true "iss matches"
  run_test 2 false "iss is invalid"
  run_test 3 false "iss is missing"

  query="$kq"
  run_test 4 true "aud matches nonce and both principals are present in keytabs"
  run_test 5 false "nonce does not match"
  run_test 6 true "aud matches nonce and single principal is in keytabs"
  run_test 7 false "principals are not in keytabs"
  run_test 8 false "keytabs are missing from claims"

  $failed && { err "Test(s) failed"; return 3; }
  return 0
}

function run_test() {
  local result
  result=$(opa eval -i "i${1}.json" -d policy.rego "query = $query" | jq '.result[].bindings.query')
  [ "$result" == "$2" ] && { err "Test $1: Pass"; return 0; } 
  err "Test $1: Fail : $3"
  failed=true
  return 10
}

function err() { echo "$@" 1>&2; }

main "$@"