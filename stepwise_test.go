package balenakeys

import (
	"fmt"
	"os"
	"sync"
	"testing"

	stepwise "github.com/hashicorp/vault-testing-stepwise"
	dockerEnvironment "github.com/hashicorp/vault-testing-stepwise/environments/docker"
	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/require"
)

// TestAccUserToken runs a series of acceptance tests to check the
// end-to-end workflow of the backend. It creates a Vault Docker container
// and loads a temporary plugin.
func TestAccUserToken(t *testing.T) {
	t.Parallel()
	if !runAcceptanceTests {
		t.SkipNow()
	}
	envOptions := &stepwise.MountOptions{
		RegistryName:    "balena",
		PluginType:      api.PluginTypeSecrets,
		PluginName:      "vault-plugin-secrets-balena",
		MountPathPrefix: "balenaAdmin",
	}

	roleName := "vault-stepwise-user-role"
	name := os.Getenv(envVarBalenaName)

	cred := new(string)
	stepwise.Run(t, stepwise.Case{
		Precheck:    func() { testAccPreCheck(t) },
		Environment: dockerEnvironment.NewEnvironment("balena", envOptions),
		Steps: []stepwise.Step{
			testAccConfig(t),
			testAccUserRole(t, roleName, name),
			testAccUserRoleRead(t, roleName, name),
			testAccUserCredRead(t, roleName, cred),
		},
	})
}

var initSetup sync.Once

func testAccPreCheck(t *testing.T) {
	initSetup.Do(func() {
		// Ensure test variables are set
		if v := os.Getenv(envVarBalenaName); v == "" {
			t.Skip(fmt.Printf("%s not set", envVarBalenaName))
		}
		if v := os.Getenv(envVarBalenaToken); v == "" {
			t.Skip(fmt.Printf("%s not set", envVarBalenaToken))
		}
		if v := os.Getenv(envVarBalenaURL); v == "" {
			t.Skip(fmt.Printf("%s not set", envVarBalenaURL))
		}
	})
}

func testAccConfig(t *testing.T) stepwise.Step {
	return stepwise.Step{
		Operation: stepwise.UpdateOperation,
		Path:      "config",
		Data: map[string]interface{}{
			"token": os.Getenv(envVarBalenaToken),
			"url":   os.Getenv(envVarBalenaURL),
		},
	}
}

func testAccUserRole(t *testing.T, roleName, username string) stepwise.Step {
	return stepwise.Step{
		Operation: stepwise.UpdateOperation,
		Path:      "role/" + roleName,
		Data: map[string]interface{}{
			"name":    username,
			"ttl":     "1m",
			"max_ttl": "5m",
		},
		Assert: func(resp *api.Secret, err error) error {
			require.Nil(t, resp)
			require.Nil(t, err)
			return nil
		},
	}
}

func testAccUserRoleRead(t *testing.T, roleName, username string) stepwise.Step {
	return stepwise.Step{
		Operation: stepwise.ReadOperation,
		Path:      "role/" + roleName,
		Assert: func(resp *api.Secret, err error) error {
			require.NotNil(t, resp)
			require.Equal(t, username, resp.Data["name"])
			return nil
		},
	}
}

func testAccUserCredRead(t *testing.T, roleName string, userToken *string) stepwise.Step {
	return stepwise.Step{
		Operation: stepwise.ReadOperation,
		Path:      "creds/" + roleName,
		Assert: func(resp *api.Secret, err error) error {
			require.NotNil(t, resp)
			require.NotEmpty(t, resp.Data["token"])
			*userToken = resp.Data["token"].(string)
			return nil
		},
	}
}
