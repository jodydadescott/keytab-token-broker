#!/bin/bash

#Default is stderr
#export KBRIDGE_LOG_TO="stderr"

#Default is info
#export KBRIDGE_LOG_LEVEL="info"
#export KBRIDGE_LOG_LEVEL="debug"
#export KBRIDGE_LOG_LEVEL="warn"
#export KBRIDGE_LOG_LEVEL="error"
export KBRIDGE_LOG_LEVEL="debug"

#Default is json
#export KBRIDGE_LOG_FORMAT="json"
#export KBRIDGE_LOG_FORMAT="console"
export KBRIDGE_LOG_FORMAT="console"

#Default is 60
#export KBRIDGE_NONCE_LIFETIME=60

#Default is 120
#export KBRIDGE_KEYTAB_LIFETIME=120

#Default is "" (all interfaces)
#export KBRIDGE_LISTEN=""

#Default is 0 (disabled)
#export KBRIDGE_HTTPPORT=0
export KBRIDGE_HTTPPORT=8080

#Default is 0 (disabled)
#export KBRIDGE_HTTPSPORT=0

#KBRIDGE_KEYTAB_PRINCIPALS has no default. It is not technically required
#but there should be one or more
#export KBRIDGE_KEYTAB_PRINCIPALS=""
export KBRIDGE_KEYTAB_PRINCIPALS="superman@EXAMPLE.COM"

#KBRIDGE_POLICY_QUERY and KBRIDGE_POLICY_REGO have no default and are 
#are required. KBRIDGE_POLICY_QUERY must be a valid OPA Rego query statement.
#KBRIDGE_POLICY_REGO must be a valid rego policy or a filename that contains
#a valid policy

export KBRIDGE_POLICY_QUERY="grant_new_nonce = data.kbridge.grant_new_nonce; data.kbridge.get_principals[get_principals]"
export KBRIDGE_POLICY_REGO='
package kbridge
	
default grant_new_nonce = false

grant_new_nonce {
	input.iss == "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo"
}

get_principals[grant] {
	grant := split(input.claims.service.keytab,",")
}
'

./kbridge-win-amd64 server