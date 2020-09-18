@ECHO OFF

SET KBRIDGE_NONCE_LIFETIME=60
SET KBRIDGE_POLICY_QUERY="grant_new_nonce = data.kbridge.grant_new_nonce; data.kbridge.get_principals[get_principals]"
SET KBRIDGE_POLICY_REGO=$(cat <<'EOF'
package kbridge
	
default grant_new_nonce = false

grant_new_nonce {
	input.iss == "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo"
}

get_principals[grant] {
	grant := split(input.claims.service.keytab,",")
}
EOF
)

SET KBRIDGE_KEYTAB_LIFETIME=300
SET KBRIDGE_KEYTAB_PRINCIPALS="superman@EXAMPLE.COM"
SET KBRIDGE_HTTPPORT=8080
# SET KBRIDGE_LISTEN=""
# SET KBRIDGE_HTTPSPORT=8443

# SET KBRIDGE_LOG_LEVEL="info"
SET KBRIDGE_LOG_LEVEL="debug"
# SET KBRIDGE_LOG_LEVEL="warn"
# SET KBRIDGE_LOG_LEVEL="error"

SET KBRIDGE_LOG_FORMAT="json"
# SET KBRIDGE_LOG_FORMAT="console"

SET KBRIDGE_LOG_TO="stderr"

SET KBRIDGE_LOG_TO="stderr"

start "" C:\Users\Admin\Desktop\kbridge-win-amd64