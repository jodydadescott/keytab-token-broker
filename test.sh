# export KBRIDGE_LOG_LEVEL="info" # default
# export KBRIDGE_LOG_LEVEL="debug"
# export KBRIDGE_LOG_LEVEL="warn"
# export KBRIDGE_LOG_LEVEL="error"

# export KBRIDGE_LOG_FORMAT="json" # default
# export KBRIDGE_LOG_FORMAT="console"

# export KBRIDGE_LOG_TO="stderr" # default

# export KBRIDGE_NONCE_LIFETIME=60 # default
# export KBRIDGE_KEYTAB_LIFETIME=120 # default

# export KBRIDGE_LISTEN="" # Default (all interfaces)

# export KBRIDGE_HTTPPORT=0 # default disabled
export KBRIDGE_HTTPPORT=8080

# export KBRIDGE_HTTPSPORT=0 # default disabled
# export KBRIDGE_HTTPSPORT=8443

# The following have no defaults and are required

export KBRIDGE_KEYTAB_PRINCIPALS="superman@EXAMPLE.COM"

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
