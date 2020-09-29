package kbridge

default grant_new_nonce = false
grant_new_nonce {
	input.iss == "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo"
}
get_principals[grant] {
	grant := split(input.claims.service.keytab,",")
}
