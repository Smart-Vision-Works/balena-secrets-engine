package balenakeys

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	balenaTokenType = "balena_token"
)

// balenaToken defines a secret for the balena token
type balenaToken struct {
	TokenID    string `json:"token_id"`
	Token      string `json:"token"`
	BalenaName string `json:"balenaName"`
}

// balenaToken defines a secret to store for a given role
// and how it should be revoked or renewed.
func (b *balenaBackend) balenaToken() *framework.Secret {
	return &framework.Secret{
		Type: balenaTokenType,
		Fields: map[string]*framework.FieldSchema{
			"token": {
				Type:        framework.TypeString,
				Description: "Balena API Key",
			},
			"balenaName": {
				Type:        framework.TypeString,
				Description: "Name of Token in Balena",
			},
		},
		Revoke: b.tokenRevoke,
		Renew:  b.tokenRenew,
	}
}

// tokenRevoke removes the token from the Vault storage API and calls the client to revoke the token
func (b *balenaBackend) tokenRevoke(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	client, err := b.getClient(ctx, req.Storage)
	if err != nil {
		return nil, fmt.Errorf("error getting client: %w", err)
	}

	tokenName := ""
	// We passed the token using InternalData from when we first created
	// the secret. This is because the balena API uses the exact token
	// for revocation. From a security standpoint, your target API and client
	// should use a token ID instead!
	nameRaw, ok := req.Secret.InternalData["key_name"]
	if ok {
		tokenName, ok = nameRaw.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value for tokenID in secret internal data")
		}
	}

	if err := deleteToken(ctx, client, tokenName); err != nil {
		return nil, fmt.Errorf("error revoking user token: %w", err)
	}
	return nil, nil
}

// tokenRenew calls the client to create a new token and stores it in the Vault storage API
func (b *balenaBackend) tokenRenew(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	roleRaw, ok := req.Secret.InternalData["role"]
	if !ok {
		return nil, fmt.Errorf("secret is missing role internal data")
	}

	// get the role entry
	role := roleRaw.(string)
	roleEntry, err := b.getRole(ctx, req.Storage, role)
	if err != nil {
		return nil, fmt.Errorf("error retrieving role: %w", err)
	}

	if roleEntry == nil {
		return nil, errors.New("error retrieving role: role is nil")
	}

	resp := &logical.Response{Secret: req.Secret}

	if roleEntry.TTL > 0 {
		resp.Secret.TTL = roleEntry.TTL
	}
	if roleEntry.MaxTTL > 0 {
		resp.Secret.MaxTTL = roleEntry.MaxTTL
	}

	return resp, nil
}

// createToken calls the balena client to sign in and returns a new token
func createToken(ctx context.Context, c *balenaClient, balenaName string, balenaDesc string) (*balenaToken, error) {

	type balenaBody struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	tokenID := uuid.New().String()

	if balenaName == "" {
		balenaName = tokenID
	}

	body := balenaBody{
		Name:        balenaName,
		Description: balenaDesc,
	}

	req, err := c.NewRequest(ctx, "POST", "api-key/user/full", "", body)
	var token string
	err = c.Do(req, &token)

	if err != nil {
		return nil, fmt.Errorf("error creating balena token: %w", err)
	}

	return &balenaToken{
		TokenID:    tokenID,
		Token:      token,
		BalenaName: balenaName,
	}, nil
}

// deleteToken calls the balena client to sign out and revoke the token
func deleteToken(ctx context.Context, c *balenaClient, tokenName string) error {
	type ApiKey struct {
		D []struct {
			ID          int       `json:"id"`
			CreatedAt   time.Time `json:"created_at"`
			Name        string    `json:"name"`
			Description string    `json:"description"`
			ExpiryDate  time.Time `json:"expiry_date"`
		} `json:"d"`
	}

	var key ApiKey

	req, err := c.NewRequest(ctx, "GET", fmt.Sprintf("v6/api_key?$select=id,created_at,name,description,expiry_date&$filter=(name%%20eq%%20%%27%s%%27)", tokenName), "", nil)

	err = c.Do(req, &key)

	if err != nil {
		return fmt.Errorf("error getting balena token: %w", err)
	}

	if len(key.D) > 0 {
		req, err = c.NewRequest(ctx, "DELETE", fmt.Sprintf("v6/api_key(%d)", key.D[0].ID), "", nil)

		var stat string
		err = c.Do(req, stat)

		// if err != nil {
		// 	return fmt.Errorf("error deleting balena token: %w", err)
		// }
	}

	return nil

}
