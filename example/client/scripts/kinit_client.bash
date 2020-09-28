#!/bin/bash
########################################################################################
# Example Script to get keytab from keytab server using OAUTH token
#
# Process Overview
# 1) Get a token from the token server
# 2) Using the token get a nonce from the Keytab server
# 3) Using the nonce attribute value get a new token from the token server with the
#    audience (aud) set to the nonce value. This is what prevents a replay attack.
# 4) Using the new token, the nonce value and the name of the desired principal
#    request the keytab from the Keytab server
# 5) Decode the base64file attribute from the Keytab into a file
# 6) Using the principal name attribute from the Keytab principal attribute
#    obtain a TGT from the Kerberos server. Note that the principal attribute
#    will differ from the original principal.
########################################################################################

PRINCIPAL="superman@EXAMPLE.COM"
KEYTAB_SERVER="35.153.18.49:8080"
TOKEN_SERVER="169.254.254.1"

function main() {
  which curl > /dev/null 2>&1 || { err "curl not found in path"; return 2; }
  which jq > /dev/null 2>&1 || { err "jq not found in path"; return 2; }

  [[ $PRINCIPAL ]] || { err "Missing env var PRINCIPAL"; return 2; }
  [[ $KEYTAB_SERVER ]] || { err "Missing env var KEYTAB_SERVER"; return 2; }
  [[ $TOKEN_SERVER ]] || { err "Missing env var TOKENB_SERVER"; return 2; }

  tmp=$(mktemp -d)
  trap cleanup EXIT

  local token
  local nonce

  err "Getting initial token"
  token=$(httpGet -H 'X-Aporeto-Metadata: secrets' "http://${TOKEN_SERVER}/token?type=OAUTH&audience=initial") || {
    err "Failed to get token"
    return 3
  }

  err "Getting nonce"
  nonce=$(httpGet "${KEYTAB_SERVER}"/getnonce\?bearertoken="${token}") || {
    err "Failed to get nonce"
    return 3
  }

  nonce=$(echo "$nonce" | jq -r '.value')

  err "Getting token again with aud set to nonce"
  token=$(httpGet -H 'X-Aporeto-Metadata: secrets' "http://${TOKEN_SERVER}/token?type=OAUTH&audience=${nonce}") || {
    err "Failed to get token"
    return 3
  }

  err "Getting keytab with token"
  keytab=$(httpGet "${KEYTAB_SERVER}"/getkeytab\?bearertoken="${token}"\&principal="${PRINCIPAL}") || {
    err "Failed to get keytab"
    err "Check you settings and also make sure that the server is running as a Domain Admin"
    return 3
  }

  echo "$keytab" | jq -r '.base64file' | base64 -d > "$tmp/keytab"
  local principal_alias
  principal_alias=$(echo "$keytab" | jq -r '.principal')

  /usr/bin/kinit -V -k -t "$tmp/keytab" "${principal_alias}" || return 4
  return 0
}

function httpGet() {
  local code
  cat /dev/null > "$tmp/response" || return 2
  code=$(curl -s -o "$tmp/response" -w "%{http_code}" "$@")
  local emsg
  local response
  response=$(<"$tmp/response")
  [ "$code" == "200" ] || {
    emsg=$(echo "$response" | jq -r .error)
    [[ $emsg ]] && err "Keytab server returned error: code->$code, message->$emsg"
    return 10
  }
  echo "$response"
  return 0
}

function cleanup() { [[ "$tmp" ]] && rm -rf "$tmp"; }

function err() { echo "$@" 1>&2; }

main "$@"
