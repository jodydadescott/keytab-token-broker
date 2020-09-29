# Keytab Token Broker (KTB)
Use Bearer Token (OAUTH, etc) to get Active Directory ephemeral Kerberos Keytabs.
 
## Overview
 
### Background
Microsoft Windows services such as CIFS use Active Directory (Kerberos + LDAP) to control authentication and authorization. Kerberos works by issuing time scoped tickets to users and computers that are part of the domain. When an application that is not part of the domain requires access to a domain service the standard method to accomplish this is to provide the external application with a Keytab file. The Keytab file consists of one or more principles (or users) and an encrypted password. The Keytab can be used to prove identity to a service on the domain.
 
### Problem
The problem is that keytabs are usually persisted to disk where they can be read by the application. This combined with the fact that the average lifetime of a Keytab is very long creates a security concern. One solution is to keep the Keytab in a vault and frequently refresh them. The industry has been evolving for some time now towards the use of ephemeral bearer tokens such as OAUTH. Bearer tokens have a finite lifetime and are usually not persisted to disk. Bearer tokens offer identity in the form of claims. The identity is attested to by the Issuer which can be verified with the issuers public key.
 
### Solution
Our solution is to use Tokens for Authentication and Authorization by executing an operator-defined Open Policy Agent (OPA) policy and then if authorized granting the bearer of the token a time scoped Keytab. We enforce the time scope by renewing the Keytabs when they reach expiration hence invalidating the previously granted Keytabs. The process is as follows.
1. A client application obtains a token from a token provider (Palo Alto Aporeto / OKTA / etc)
1. Using the Token the client makes a Nonce request to the KTB server
1. The KTB server validates the Token, executes an OPA policy to authorize the request, and returns the Nonce (JSON with fields value and exp)
1. The client makes a request for a new Token from the token provider with the aforementioned Nonce as the Audience (aud) field
1. Using the Token the client makes a request to the KTB server for a Keytab for a desired Principal
1. The KTB server verifies the Token, Nonce and executes an OPA policy to authorize the request and returns the Keytab (JSON with fields base64file, softExp and hardExp)
1. The client extracts the Base64 from the Keytab JSON object and recreates a Keytab file
 
### Caveats
Unlike tokens, there can only be one valid Keytab at a time. Consider the situation where a Keytab is requested and an expiration of 120 seconds in the future is assigned. Then consider that 110 seconds later a new Keytab is requested. If we hand out the existing Keytab it only has 10 seconds of life left. If we recreate a new Keytab then we are breaking the contract for the first Keytab. Our solution is to reset the expiration time in this situation. This creates a problem that if a key is continuously requested before expiration it will never really expire. Our solution for this is to have a soft expiration and a hard expiration. The soft expiration can be incremented but the hard expiration cannot be increased. We expose both the soft and hard expiration values as fields in the request.
 
The reason we do not issue TGT tickets is that it is difficult to install these into the local cache. Linux has several different implementations of ticket caches and Windows does not, AFAIK, expose an API to read/write the TGT cache. Using the Keytab format also makes it easier for an existing script that expects a keytab file.
 
## Notes
1. KTB is designed to run on Windows. It will run on Linux and Darwin but it will create dummy keytabs. This may be useful for testing functionality.
1. It uses the utility C:/Windows/System32/ktpass to create Keytabs.

## Installation and Configuration

### Installation
1. Create a directory `mkdir C:\Program Files\KTBServer`
1. Download the ktbserver-windows-amd64.exe binary and place in the newly created directory
1. Create an example configuration with the command `.\ktbserver-windows-amd64.exe config example > config.yaml`
1. Install as a Windows service with `.\ktbserver.exe service install`
1. Configure the service to run as a Domain Admin. If you don’t do this it will NOT be able to create Keytabs.
1. Start the service with `.\ktbserver-windows-amd64.exe service start` or use the Windows Service utility

## Building the configuration (OPA / Rego)
Authorization is done with OPA. You will want to write a custom policy. This is the example one found in the example configuration.
```
package kbridge

default grant_new_nonce = false
grant_new_nonce {
	input.iss == "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo"
}
get_principals[grant] {
	grant := split(input.claims.service.keytab,",")
}
```
Save your config and name it policy.rego. Then create an input configuration file with the name input.yaml similiar to this.
```
apiVersion: V1
network:
  Listen: any
  httpPort: 8080
  httpsPort: 8443
policy:
  # query: grant_new_nonce = data.kbridge.grant_new_nonce; data.kbridge.get_principals[get_principals]
  query: {your query goes here}
  nonceLifetime: 60
  keytabSoftLifetime: 120
  keytabHardLifetime: 600
logging:
  logLevel: info
  logFormat: json
  outputPaths:
  - stderr
  errorOutputPaths:
  - stderr
data:
  principals:
  - user1@EXAMPLE.COM
  - user2@EXAMPLE.COM
```
Be sure to update the OPA/Rego query with your query and your users. Then use the command below to build a single configuration file
```
.\ktbserver-windows-amd64.exe config make --config input.yaml,policy.rego > config.yaml
```
Note that configuration files are processed in order and that subsequent configurations will overwrite existing settings if the attribute is set.

## Example(s)
 
### Example Client
[Example Client kinit wrapper](example/client/scripts/kinit_client.bash)