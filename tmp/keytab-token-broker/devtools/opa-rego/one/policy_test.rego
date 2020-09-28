package authz

test_1 {
  allow with input as {"iss":"https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo"}
}

test_2 {
  not allow with input as {"iss":"evil"}
}
