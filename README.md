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
1. The KTB server verifies the Token, Nonce and executes an OPA policy to authorize the request and returns the Keytab (JSON with fields base64file, and exp)
1. The client extracts the Base64 from the Keytab JSON object and recreates a Keytab file

## Details

### Keytabs
Keytabs hold one or more principals or users and their corresponding password. The password is encrypted by the Kerberos Domain Controller (KDC). The Keytab is valid as long as the corresponding principal account is valid and the password is not changed. For our implementation we only support one principal at this time.

### Keytab Lifetime
The keytab lifetime is set by runtime configuration. At the top of each lifetime period a new keytab will be generated for each configured principal. For example if the lifetime 


Our implementation only supports one principal. The principal is the username + the domain with "HTTP/" prepended. For example the user "superman" on the domain "EXAMPLE.COM" will have the principal name of "HTTP/superman@EXAMPLE.COM". The lifetime of keytabs are set by configuration at runtime. At the top of the lifetime keytabs are created for each principal in the runtime configuration. The password is created using a One Time Password (OTP) from a OTP generator combined with the principal and the OTP seed. The OTP seed is set in the configuration. This means that the server can be ran on multiple servers with no state coordination as long as the seed matches. Obviously the seed must be kept secure.

### Caveats

Keytabs hold encrypted passwords. They are invalidated when the password is changed and AFAIK Windows will only support one password at a time. Hence their can only be 

Unlike tokens, there can only be one valid Keytab at a time. Consider the situation where a Keytab is requested and an expiration of 120 seconds in the future is assigned. Then consider that 110 seconds later a new Keytab is requested. If we hand out the existing Keytab it only has 10 seconds of life left. If we recreate a new Keytab then we are breaking the contract for the first Keytab. Our solution is to reset the expiration time in this situation. This creates a problem that if a key is continuously requested before expiration it will never really expire. Our solution for this is to have a soft expiration and a hard expiration. The soft expiration can be incremented but the hard expiration cannot be increased. We expose both the soft and hard expiration values as fields in the request.

KTB is intended to run on Windows and have access to the binary "C:/Windows/System32/ktpass". It is possible to run on Linux and Darwin but valid Keytabs will not be created. This is useful for testing functionality such as OPA where a Windows server and domain are not readibly avaiable.

## Installation and Configuration

### Requirements

1. A Windows server that is part of an Active Directory Domain and Directory Admin privileges
1. A Token provider and method to obtain tokens

### Installation

1. Create a directory `mkdir "C:\Program Files\KTBServer"`
1. Download the ktbserver-windows-amd64.exe binary and place in the newly created directory
1. Install as a Windows service with `.\ktbserver-windows-amd64.exe service install`
1. Configure the service to run as a Directory Administrator by going to services, click on service, select properties from the menu, select the tab "Log On", select the button "This account:", enter the username and password for a directory admin, and press save.

### Configuration

Create an example configuration with the command `.\ktbserver-windows-amd64.exe config example > config.yaml`. You will then need to edit the configuration for your environment. You can edit the OPA policy directly in the config or you can create a seperate file such as policy.rego. Configurations can be merged together with the command `.\ktbserver-windows-amd64.exe config make --config input.yaml,policy.rego > config.yaml`. Note that configuration files are processed in order and that subsequent configurations will overwrite existing settings if the attribute is set. Once you have the config created store it on the local file system or in a repo and use the command `.\ktbserver-windows-amd64.exe config set $location` where \$location is the file or URL of the configuration.

### Start

Start the service with the command `.\ktbserver-windows-amd64.exe service start` or use the Windows service manager to start.

## Examples

### Example OPA/rego policy and query statement
This policy grants nonce privileges to those bearing a token with the matching issuer and returns a list of authorized principals by extracting the claim service.keytab from the token and splitting it on commas.
1. For nonce bearer token must have matching iss (Issuer)
1. Have a matching Issuer (iss) and matching principal in the claim "keytab" where principals are seperated with commas to get a keytab

```
package kbridge

default grant_new_nonce = false
grant_new_nonce {
	input.iss == "https://..."
}
get_principals[grant] {
	grant := split(input.claims.service.keytab,",")
}
```

And the corresponding query

```
"grant_new_nonce = data.kbridge.grant_new_nonce; data.kbridge.get_principals[get_principals]"
```

### Example Client

[Example Client kinit wrapper](example/client/scripts/kinit_client.bash)

## FAQ

The reason we do not issue TGT tickets is that it is difficult to install these into the local cache. Linux has several different implementations of ticket caches and Windows does not, AFAIK, expose an API to read/write the TGT cache. Using the Keytab format also makes it easier for an existing script that expects a keytab file.
