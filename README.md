# vault-plugin-secrets-balena

This is a Vault plugin to manage API Keys for the IoT Platform [Balena](https://www.balena.io/)

## Try it out!

You can run a set of commands to enable the secrets engine at `/balena` in
Vault.

Then, you can write a configuration and a role based on a Balena account.

NOTE: Each role requires a different Balena account and balenaApiKey (Session Token) associated with that account

Finally, you can read the credentials for the role.

```shell
$ CGO_ENABLED=0 go build -ldflags="-extldflags=-static" -o vault/plugins/vault-plugin-secrets-balena cmd/vault-plugin-secrets-balena/main.go
```

Once built, copy the binary to your vault plugins folder. Then run the following to active the engine:

```shell
$ SHA256=$(sha256sum vault/plugins/vault-plugin-secrets-balena | cut -d ' ' -f1)

vault plugin register -sha256=$SHA256 secret vault-plugin-secrets-hashicups
Success! Registered plugin: vault-plugin-secrets-balena

vault secrets enable -path=balena vault-plugin-secrets-balena
Success! Enabled the vault-plugin-secrets-balena secrets engine at: balena/

vault write balena/config url="https://api.balena-cloud.com"
Success! Data written to: balena/config

vault write balen/role/developer balenaApiKey="${BALENA_SESSION_TOKEN}" ttl="5m" max_ttl="1h"
Success! Data written to: balena/role/developer

vault read balena/creds/developer
Key                Value
---                -----
lease_id           balena/creds/default/tVsj1JusAp8mW2vgD3FqAnxf
lease_duration     5m
lease_renewable    true
key_desc           this is a test token managed by Vault
key_name           test-balena-apikey
token              Aej6vxnlTA4ifgH8Ak16Jtj8oGjjlALQ
token_id           5f83a6ee-3b51-44e4-9744-76e467762fde
```

![Alt text](https://storage.googleapis.com/static_assets/scripts/balena-key-example.png)

Copy the token and set it to the `TOKEN` environment variable.

```shell
export TOKEN="Bearer Aej6vxnlTA4ifgH8Ak16Jtj8oGjjlALQ"
```

Call the Balena API to test the token.

```shell
$ curl -i -X GET -H "Authorization: ${TOKEN}" -H  "Content-Type: application/json" https://api.balena-cloud.com/user/v1/whoami

TTP/2 200 
date: Mon, 25 Sep 2023 15:01:50 GMT
content-type: application/json; charset=utf-8
content-length: 70
etag: W/"46-n2a7afpiWLDYhYthxsDKLupNKGg"
vary: Accept-Encoding
cf-cache-status: DYNAMIC
strict-transport-security: max-age=15552000
server: cloudflare
cf-ray: 80c4250e0fa6ec80-SEA
alt-svc: h3=":443"; ma=86400

{"id":54623,"username":"developer_87","email":"developer@mydomain.com"}
```

Revoke the lease for the Balena token in Vault.

```shell
$ vault lease revoke balena/creds/developer/tVsj1JusAp8mW2vgD3FqAnxf

All revocation operations queued successfully!
```

If you try to call the Balena API again, you'll find that the token is no longer valid.

```shell
$ curl -i -X GET -H "Authorization: ${TOKEN}" -H  "Content-Type: application/json" https://api.balena-cloud.com/user/v1/whoami

HTTP/2 401 
date: Mon, 25 Sep 2023 15:03:01 GMT
cf-cache-status: DYNAMIC
strict-transport-security: max-age=15552000
server: cloudflare
cf-ray: 80c426cbac1f2841-SEA
alt-svc: h3=":443"; ma=86400

```

## Additional references:

- [Upgrading Plugins](https://www.vaultproject.io/docs/upgrading/plugins)
- [List of Vault Plugins](https://www.vaultproject.io/docs/plugin-portal)

## FAQ

### Session Tokens

Currently API keys generated for an account cannot generate new API Keys, so you have to use the Session Token when configuring the role. It expires every 7 days so must be rotated. You can automate this with scripting.

![Alt text](https://storage.googleapis.com/static_assets/scripts/balena-session-token.png)