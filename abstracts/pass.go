package abstracts

import "github.com/Lukiya/oauth2go"

type (
	AuthServerOptions struct {
		BaseServerOptions
		oauth2go.AuthServerOptions
		PrivateKeyPath string
		HashKey        string
		BlockKey       string
		ClientStoreKey string
		TokenStoreKey  string
		WebRoot        string
	}
)
