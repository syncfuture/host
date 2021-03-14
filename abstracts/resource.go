package abstracts

import "github.com/Lukiya/oauth2go/model"

type (
	OAuthResourceOptions struct {
		IrisBaseServerOptions
		PublicKeyPath    string
		SigningAlgorithm string
		OAuth            *model.Resource
	}
)
