# Keytab Token Broker (KTB)
Use Bearer Token (OAUTH, etc) to get Active Directory ephemeral Kerberos Keytabs.

## Overview

### Background
Microsoft Windows services such as CIFS use Active Directory (Kerberos + LDAP) to control authenticaton and authorization. Kerberos works by issuing time scoped tickets to users and computers that are part of the domain. When an application that is not part of the domain requires access to a domain service the standard method to accompolish this is to provide the external application with a Keytab file. The Keytab file consist of one or more principlas (or users) and an encrypted password. The Keytab can be used to prove identity to a service on the domain.

### Problem
The problem is that keytabs are usually persisted to disk where they can be read by the application. This combined with the fact that the average lifetime of a Keytab is very long creates a security concern. One solution is to keep the Keytab in a vault and frequently refresh them. The industry has been evolving for some time now towards the use of ephemeral bearer tokens such as OAUTH. Bearer tokens have finite lifetime and are usually not persisted to disk. Bearer tokens offer identity in the form of claims. The identity is attested to by the Issuer which can be verified with the issuers public key.

### Solution
Our solution is to use Tokens for Authentication and Authorization by executing a operator defined Open Policy Agent (OPA) policy and then if authorized granting the bearer of the token a time scoped Keytab. We enforce the time scope by renewing the Keytabs when they reach expiration hence invalidating the previously granted Keytabs. The process is as follows.
1. A client application obtains a token from a token provider (Palo Alto Aporeto / OKTA / etc)
1. Using the Token the client makes a Nonce request to the KTB server
1. The KTB server validates the Token, executes an OPA policy to authorize the request, and returns the Nonce (JSON with fields value and exp)
1. The client makes a request for a new Token from the token provider with the aforementioned Nonce as the Audience (aud) field
1. Using the Token the cliet makes a request to the KTB server for a Keytab for a desired Principal
1. The KTB server verifies the Token, Nonce and executes an OPA policy to authorize the request and returns the Keytab (JSON with fields base64file, softExp and hardExp)
1. The client extracts the Base64 from the Keytab JSON object and recreates a Keytab file

### Caveats
Unlike tokens there can only be one valid Keytab at a time. Consider the situation where a Keytab is requested and an expiration of 120 seconds in the future is assigned. Then consider that 110 seconds later a new Keytab is requested. If we hand out the existing Keytab it only has 10 seconds of life left. If we recreate a new Keytab then we are breaking the contract for the first Keytab. Our solution is to reset the expiration time in this situation. This creates a problem that if a key is continously requested before expiration it will never really expire. Our solution for this is to have a soft expiration and a hard expiration. The soft expiration can be incremented but the hard expiration cannot be increased. We expose both the soft and hard expiration values as fields in the request.

## Notes
KTB is designed to run on Windows. It uses the utility C:/Windows/System32/ktpass to create Keytabs. For testing purposes KTB may be ran on Linux or Darwin. In this situation only dummy Keytabs will be provided. The base64 will NOT be a valid Keytab file.

## Example(s)

### Example Client
[Example Client Script](example/client/scripts/kinit_client.bash)

## Installation -- WORK Needed
1. Create the directory C:\Program Files\KTBServer
1. Download the ktbserver.exe binary to C:\Program Files\KTBServer/ktbserver.exe
1. Create and edit the configuration. 
1. Edit the configuration and place it somewhere on the local disk or in a Git Repo
1. Set the config location with the command `C:\Program Files\KTBServer/ktbserver.exe config set LOCATION`. LOCATION can be a URL or local file.
1. Install the
Download the ktbserver.exe binary to a Windows domain controller or member. 

`C:\Program Files\KTBServer/ktbserver.exe config example > ktbserver.conf`


