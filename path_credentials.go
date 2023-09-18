package balenakeys

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// pathCredentials extends the Vault API with a `/creds`
// endpoint for a role. You can choose whether
// or not certain attributes should be displayed,
// required, and named.
func pathCredentials(b *balenaBackend) *framework.Path {
	return &framework.Path{
		Pattern: "creds/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeLowerCaseString,
				Description: "Name of the role",
				Required:    true,
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation:   b.pathCredentialsRead,
			logical.UpdateOperation: b.pathCredentialsRead,
		},

		HelpSynopsis:    pathCredentialsHelpSyn,
		HelpDescription: pathCredentialsHelpDesc,
	}
}

// pathCredentialsRead creates a new balena token each time it is called if a
// role exists.
func (b *balenaBackend) pathCredentialsRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	roleName := d.Get("name").(string)

	roleEntry, err := b.getRole(ctx, req.Storage, roleName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving role: %w", err)
	}

	if roleEntry == nil {
		return nil, errors.New("error retrieving role: role is nil")
	}

	if roleEntry.Name != "" {
		return b.createUserCreds(ctx, req, roleEntry)
	}

	resp := &logical.Response{
		Data: map[string]interface{}{
			"token":    roleEntry.Token,
			"token_id": roleEntry.TokenID,
			"role":     roleEntry.Name,
		},
	}
	return resp, nil
}

// createUserCreds creates a new balena token to store into the Vault backend, generates
// a response with the secrets information, and checks the TTL and MaxTTL attributes.
func (b *balenaBackend) createUserCreds(ctx context.Context, req *logical.Request, role *balenaRoleEntry) (*logical.Response, error) {
	token, err := b.createToken(ctx, req.Storage, role)
	if err != nil {
		return nil, err
	}

	// The response is divided into two objects (1) internal data and (2) data.
	// If you want to reference any information in your code, you need to
	// store it in internal data!
	resp := b.Secret(balenaTokenType).Response(map[string]interface{}{
		"token":    token.Token,
		"token_id": token.TokenID,
	}, map[string]interface{}{
		"token": token.Token,
		"role":  role.Name,
	})

	if role.TTL > 0 {
		resp.Secret.TTL = role.TTL
	}

	if role.MaxTTL > 0 {
		resp.Secret.MaxTTL = role.MaxTTL
	}

	return resp, nil
}

// createToken uses the balena client to sign in and get a new token
func (b *balenaBackend) createToken(ctx context.Context, s logical.Storage, roleEntry *balenaRoleEntry) (*balenaToken, error) {
	client, err := b.getClient(ctx, s)
	if err != nil {
		return nil, err
	}

	var token *balenaToken

	token, err = createToken(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("error creating balena token: %w", err)
	}

	if token == nil {
		return nil, errors.New("error creating balena token")
	}

	return token, nil
}

const pathCredentialsHelpSyn = `
Generate a balena API token from a specific Vault role.
`

const pathCredentialsHelpDesc = `
This path generates a balena API user tokens
based on a particular role. A role can only represent a user token,
since balena doesn't have other types of tokens.
`
