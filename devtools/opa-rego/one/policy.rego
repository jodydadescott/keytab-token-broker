package authz

default allow = false

allow = true {
  input.iss == "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo"
}
