package balenakeys

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
)

const (
	envVarRunAccTests = "VAULT_ACC"
	envVarBalenaToken = "TEST_BALENA_TOKEN"
	envVarBalenaURL   = "TEST_BALENA_URL"
	envVarBalenaName  = "TEST_BALENA_NAME"
)

// getTestBackend will help you construct a test backend object.
// Update this function with your target backend.
func getTestBackend(tb testing.TB) (*balenaBackend, logical.Storage) {
	tb.Helper()

	config := logical.TestBackendConfig()
	config.StorageView = new(logical.InmemStorage)
	config.Logger = hclog.NewNullLogger()
	config.System = logical.TestSystemView()

	b, err := Factory(context.Background(), config)
	if err != nil {
		tb.Fatal(err)
	}

	return b.(*balenaBackend), config.StorageView
}

// runAcceptanceTests will separate unit tests from
// acceptance tests, which will make active requests
// to your target API.
var runAcceptanceTests = os.Getenv(envVarRunAccTests) == "1"

// testEnv creates an object to store and track testing environment
// resources
type testEnv struct {
	Token string
	URL   string
	Name  string

	Backend logical.Backend
	Context context.Context
	Storage logical.Storage

	// SecretToken tracks the API token, for checking rotations
	SecretToken string

	// Tokens tracks the generated tokens, to make sure we clean up
	Tokens []string
}

// AddConfig adds the configuration to the test backend.
// Make sure data includes all of the configuration
// attributes you need and the `config` path!
func (e *testEnv) AddConfig(t *testing.T) {
	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "config",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"url": e.URL,
		},
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	require.Nil(t, resp)
	require.Nil(t, err)
}

// AddUserTokenRole adds a role for the HashiCups
// user token.
func (e *testEnv) AddUserTokenRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "role/test-user-token",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"name":         e.Name,
			"balenaApiKey": e.Token,
		},
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	require.Nil(t, resp)
	require.Nil(t, err)
}

// ReadUserToken retrieves the user token
// based on a Vault role.
func (e *testEnv) ReadUserToken(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "creds/test-user-token",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"ttl": "3600h",
		},
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	require.Nil(t, err)
	require.NotNil(t, resp)

	if t, ok := resp.Data["token_id"]; ok {
		e.Tokens = append(e.Tokens, t.(string))
	}
	require.NotEmpty(t, resp.Data["token_id"])

	if e.SecretToken != "" {
		require.NotEqual(t, e.SecretToken, resp.Data["token_id"])
	}

	// collect secret IDs to revoke at end of test
	require.NotNil(t, resp.Secret)
	if t, ok := resp.Secret.InternalData["token_id"]; ok {
		e.SecretToken = t.(string)
	}
}

// CleanupUserTokens removes the tokens
// when the test completes.
// func (e *testEnv) CleanupUserTokens(t *testing.T) {
// 	if len(e.Tokens) == 0 {
// 		t.Fatalf("expected 2 tokens, got: %d", len(e.Tokens))
// 	}

// 	for _, token := range e.Tokens {
// 		b := e.Backend.(*balenaBackend)
// 		client, err := b.getClient(e.Context, e.Storage)
// 		if err != nil {
// 			t.Fatal("fatal getting client")
// 		}

// 		if err := deleteToken(e.Context, client, token); err != nil {
// 			t.Fatalf("unexpected error deleting user token: %s", err)
// 		}
// 	}
// }
