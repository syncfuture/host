package abstracts

import (
	"crypto/rsa"

	"github.com/Lukiya/oauth2go/model"
	"github.com/pascaldekloe/jwt"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/srsautil"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host/shttp"
)

type (
	OAuthResourceHost struct {
		*BaseHost
		OAuthOptions     *model.Resource `json:"OAuth,omitempty"`
		PublicKeyPath    string
		SigningAlgorithm string
		PublicKey        *rsa.PublicKey
		authMiddleware   IAuthMiddleware
	}
)

func (x *OAuthResourceHost) BuildOAuthResourceHost() {
	x.BaseHost.BuildBaseHost()

	if x.OAuthOptions == nil {
		log.Fatal("OAuth secion in configuration is missing")
	}

	if x.PublicKeyPath == "" {
		log.Fatal("public key path cannot be empty")
	}
	if x.OAuthOptions == nil {
		log.Fatal("oauth options cannot be nil")
	}
	if x.OAuthOptions.ValidIssuers == nil || len(x.OAuthOptions.ValidIssuers) == 0 {
		log.Fatal("Issuers cannot be empty")
	}
	if x.OAuthOptions.ValidAudiences == nil || len(x.OAuthOptions.ValidAudiences) == 0 {
		log.Fatal("Audiences cannot be empty")
	}
	if x.SigningAlgorithm == "" {
		x.SigningAlgorithm = jwt.PS256
	}

	if x.URLProvider != nil {
		for i := range x.OAuthOptions.ValidIssuers {
			x.OAuthOptions.ValidIssuers[i] = x.URLProvider.RenderURL(x.OAuthOptions.ValidIssuers[i])
		}
		for i := range x.OAuthOptions.ValidAudiences {
			x.OAuthOptions.ValidAudiences[i] = x.URLProvider.RenderURL(x.OAuthOptions.ValidAudiences[i])
		}
	}

	// read public certificate
	cert, err := srsautil.ReadCertFromFile(x.PublicKeyPath)
	u.LogFaltal(err)
	x.PublicKey = cert.PublicKey.(*rsa.PublicKey)

	////////// auth middleware
	if x.authMiddleware == nil {
		x.authMiddleware = newJWTAuthMiddleware(x.PublicKey, x.OAuthOptions.ValidAudiences, x.OAuthOptions.ValidIssuers, nil, x.PermissionAuditor)
	}
}

func (x *OAuthResourceHost) Auth(next shttp.RequestHandler, routes ...string) shttp.RequestHandler {
	return x.authMiddleware.Serve(next, routes...)
}
