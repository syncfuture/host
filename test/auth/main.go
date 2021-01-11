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
		oauth2core.Claim_Name:    username,
		oauth2core.Claim_Issuer:  "https://p.test.com",
		oauth2core.Claim_Expire:  exp,
		oauth2core.Claim_IssueAt: utcNow.Unix(),
		// oauth2core.Claim_NotValidBefore: utcNow.Unix(),
	}

	rexp := client.GetRefreshTokenExpireSeconds()
	if rexp > 0 {
		r[oauth2core.Claim_RefreshTokenExpire] = utcNow.Add(time.Duration(rexp) * time.Second).Unix()
	}

	r[oauth2core.Claim_Audience] = client.GetAudiences()
	r[oauth2core.Claim_Scope] = scopes

	if grantType == oauth2core.GrantType_Client {
		r[oauth2core.Claim_Name] = client.GetID()
		r[oauth2core.Claim_Role] = "1"
	} else {
		r[oauth2core.Claim_Subject] = "123456789"
		r[oauth2core.Claim_Email] = "test@test.com"
		r[oauth2core.Claim_Role] = "4"
		r[oauth2core.Claim_Level] = "5"
		r[oauth2core.Claim_Status] = "1"
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
