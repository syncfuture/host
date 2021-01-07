//go:generate go get -u github.com/valyala/quicktemplate/qtc
//go:generate qtc -dir=./...
package main

import (
	"time"

	oauth2core "github.com/Lukiya/oauth2go/core"
	"github.com/Lukiya/oauth2go/model"
	"github.com/Lukiya/oauth2go/security"
	"github.com/Lukiya/oauth2go/token"
	"github.com/syncfuture/host"
	"github.com/valyala/fasthttp"
)

func newClaimsGenerator() token.ITokenClaimsGenerator {
	return &myClaimsGenerator{}
}

type myClaimsGenerator struct{}

func (x *myClaimsGenerator) Generate(ctx *fasthttp.RequestCtx, grantType string, client model.IClient, scopes []string, username string) *map[string]interface{} {
	utcNow := time.Now().UTC()
	exp := utcNow.Add(time.Duration(client.GetAccessTokenExpireSeconds()) * time.Second).Unix()

	r := map[string]interface{}{
		"name": username,
		"iss":  "https://p.test.com",
		"exp":  exp,
		"iat":  utcNow.Unix(),
		"nbf":  utcNow.Unix(),
	}

	r["aud"] = client.GetAudiences()
	r["scope"] = scopes

	if grantType == oauth2core.GrantType_Client {
		r["name"] = client.GetID()
		r["role"] = "1"
	} else {
		r["sub"] = "123456789"
		r["name"] = username
		r["email"] = "test@test.com"
		r["role"] = "4"
		r["level"] = "5"
		r["status"] = "1"
	}

	return &r
}

func newResourceOwnerValidator() security.IResourceOwnerValidator {
	return &myResourceOwnerValidator{}
}

type myResourceOwnerValidator struct{}

func (x *myResourceOwnerValidator) Verify(username, password string) bool {
	return username == password
}

var (
	_options    *host.AuthServerOptions
	_authServer *host.OAuthServer
)

func main() {
	_options = host.NewAuthServerOptions()
	_options.ClaimsGenerator = newClaimsGenerator()
	_options.ResourceOwnerValidator = newResourceOwnerValidator()
	_authServer = host.NewOAuthServer(_options)

	// home
	_authServer.Get("/", homePageGet)
	// login
	_authServer.Get(_options.LoginEndpoint, loginPageGet)
	_authServer.Post(_options.LoginEndpoint, loginPagePost)

	_authServer.Run()
}
