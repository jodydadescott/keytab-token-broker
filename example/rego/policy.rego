package kbridge

# Input data format
# {
#   "claims": {
#     "alg": "EC",
#     "kid": "donut",
#     "iss": "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo",
#     "exp": 1599844897,
#     "aud": "daisy",
#     "service": {
#       "keytab": "user1@example.com,user2@example.com"
#     }
#   },
#   "principal": "user1@example.com",
#   "nonce": "daisy"
# }

default auth_get_nonce = false
default auth_get_keytab = false

auth_base {
   # Match Issuer
   input.claims.iss == "abc123"
}

auth_get_nonce {
   auth_base
}

auth_nonce {
   # Verify that the request nonce matches the expected nonce. Our token provider
   # has the nonce in the audience field under claims
   input.claims.aud == input.nonce
}

auth_get_keytab {
   # The nonce must be validated and then the principal. This is done by splitting the
   # principals in the claim service.keytab by the comma into a set and checking for
   # match with requested principal
   auth_base
   auth_nonce
   split(input.claims.service.keytab,",")[_] == input.principal
}