package authz

default allow = false

allow {
	input.iss == "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo"
	input.service.ilove == "the80s"
}

user_is_granted[grant] {
	grant := input.service.keytab
}
