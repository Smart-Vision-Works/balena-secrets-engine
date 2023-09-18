# vault-plugin-secrets-balena

## Try it out!

You can run a set of commands to enable the secrets engine at `/hashicups` in
Vault.

Then, you can write a configuration and a `test` role based on a HashiCups username.

Finally, you can read the credentials for the `test` role.

```shell
$ make vault_plugin

vault secrets enable -path=hashicups vault-plugin-secrets-hashicups
Success! Enabled the vault-plugin-secrets-hashicups secrets engine at: hashicups/
vault write hashicups/config username="vault-plugin-testing" password='Testing!123' url="${TEST_HASHICUPS_URL}"
Success! Data written to: hashicups/config
vault write hashicups/role/test username="vault-plugin-testing"
Success! Data written to: hashicups/role/test
vault read hashicups/creds/test
Key                Value
---                -----
lease_id           hashicups/creds/test/tVsj1JusAp8mW2vgD3FqAnxf
lease_duration     768h
lease_renewable    true
token              eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjY5MDI1MzQsInRva2VuX2lkIjoyNywidXNlcl9pZCI6MSwidXNlcm5hbWUiOiJ2YXVsdC1wbHVnaW4tdGVzdGluZyJ9.ZlH4ysV3860KbqU-rZHeQJ8p_WT6TCNrr_rWB075efY
token_id           5f83a6ee-3b51-44e4-9744-76e467762fde
user_id            1
username           vault-plugin-testing
```

Copy the token and set it to the `TOKEN` environment variable.

```shell
export TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjY5MDI1MzQsInRva2VuX2lkIjoyNywidXNlcl9pZCI6MSwidXNlcm5hbWUiOiJ2YXVsdC1wbHVnaW4tdGVzdGluZyJ9.ZlH4ysV3860KbqU-rZHeQJ8p_WT6TCNrr_rWB075efY
```

Call the HashiCups API to create a new coffee product. You should successfully create a new Melbourne Magic coffee offering.

```shell
$ curl -i -X POST -H "Authorization:${TOKEN}" ${TEST_HASHICUPS_URL}/coffees -d '{"name":"melbourne magic", "teaser": "delicious custom coffee", "description": "best coffee in the world"}'

HTTP/1.1 200 OK
Date: Tue, 20 Jul 2021 21:25:38 GMT
Content-Length: 87
Content-Type: text/plain; charset=utf-8

{"id":9,"name":"","teaser":"","description":"","price":0,"image":"","ingredients":null}
```

Revoke the lease for the HashiCups token in Vault.

```shell
$ vault lease revoke hashicups/creds/test/tVsj1JusAp8mW2vgD3FqAnxf

All revocation operations queued successfully!
```

If you try to add a new coffee product, tonic espresso, to HashiCups, you'll find that the token is no longer valid.

```shell
$ curl -i -X POST -H "Authorization:${TOKEN}" ${TEST_HASHICUPS_URL}/coffees -d '{"name":"tonic espresso", "teaser": "delicious custom coffee", "description": "best coffee in the world"}'

HTTP/1.1 401 Unauthorized
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Tue, 20 Jul 2021 21:27:47 GMT
Content-Length: 14

Invalid token
```

## Additional references:

- [Upgrading Plugins](https://www.vaultproject.io/docs/upgrading/plugins)
- [List of Vault Plugins](https://www.vaultproject.io/docs/plugin-portal)