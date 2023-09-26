package balenakeys

import (
	"context"
	"strings"
	"sync"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// Factory returns a new backend as logical.Backend
func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := backend()
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

// balenaBackend defines an object that
// extends the Vault backend and stores the
// target API's client.
type balenaBackend struct {
	*framework.Backend
	lock   sync.RWMutex
	client *balenaClient
}

// backend defines the target API backend
// for Vault. It must include each path
// and the secrets it will store.
func backend() *balenaBackend {
	var b = balenaBackend{}

	b.Backend = &framework.Backend{
		Help: strings.TrimSpace(backendHelp),
		PathsSpecial: &logical.Paths{
			LocalStorage: []string{
				// WAL stands for Write-Ahead-Log, which is used for Vault replication
				framework.WALPrefix,
			},
			SealWrapStorage: []string{
				"config",
				"role/*",
			},
		},
		Paths: framework.PathAppend(
			pathRole(&b),
			[]*framework.Path{
				pathConfig(&b),
				pathCredentials(&b),
			},
		),
		Secrets: []*framework.Secret{
			b.balenaToken(),
		},
		BackendType: logical.TypeLogical,
		Invalidate:  b.invalidate,
	}
	return &b
}

// reset clears any client configuration for a new
// backend to be configured
func (b *balenaBackend) reset() {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.client = nil
}

// invalidate clears an existing client configuration in
// the backend
func (b *balenaBackend) invalidate(ctx context.Context, key string) {
	if key == "config" {
		b.reset()
	}
}

// getClient locks the backend as it configures and creates a
// a new client for the target API
func (b *balenaBackend) getClient(ctx context.Context, s logical.Storage, bToken string) (*balenaClient, error) {
	b.lock.RLock()
	unlockFunc := b.lock.RUnlock
	defer func() { unlockFunc() }()

	b.lock.RUnlock()
	b.lock.Lock()
	unlockFunc = b.lock.Unlock

	config, err := getConfig(ctx, s)
	if err != nil {
		return nil, err
	}

	// if b.client == nil {
	if config == nil {
		config = new(balenaConfig)
	}
	// }

	b.client, err = newClient(config, bToken)
	if err != nil {
		return nil, err
	}

	return b.client, nil
}

// backendHelp should contain help information for the backend
const backendHelp = `
The balena secrets backend dynamically generates user tokens.
After mounting this backend, credentials to manage balena user tokens
must be configured with the "config/" endpoints.
`
