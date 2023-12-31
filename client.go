package balenakeys

import (
	"errors"

	"go.einride.tech/balena"
)

// balenaClient creates an object storing
// the client.
type balenaClient struct {
	*balena.Client
}

// newClient creates a new client to access balena
// and exposes it for any secrets or roles to use.
func newClient(config *balenaConfig, bToken string) (*balenaClient, error) {
	if config == nil {
		return nil, errors.New("client configuration was nil")
	}

	// if config.Token == "" {
	// 	return nil, errors.New("client token was not defined")
	// }

	if config.URL == "" {
		return nil, errors.New("client URL was not defined")
	}

	c := balena.New(nil, bToken)

	return &balenaClient{c}, nil
}
