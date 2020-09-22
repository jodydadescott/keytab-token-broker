@ECHO OFF

SET KBRIDGE_LOGTO="C:\Users\Admin\Desktop\kbridg\log.txt"
SET KBRIDGE_LOGLEVEL="debug"
SET KBRIDGE_LOGFORMAT="console"
SET KBRIDGE_HTTPPORT=8080
SET KBRIDGE_KEYTABPRINCIPALS="superman@EXAMPLE.COM"

SET KBRIDGE_POLICYQUERY="grant_new_nonce = data.kbridge.grant_new_nonce; data.kbridge.get_principals[get_principals]"
SET KBRIDGE_POLICYREGO='
package kbridge
	
default grant_new_nonce = false

grant_new_nonce {
	input.iss == "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo"
}

get_principals[grant] {
	grant := split(input.claims.service.keytab,",")
}
'

start "" C:\Users\Admin\Desktop\kbridge\kbridge-win-amd64 server