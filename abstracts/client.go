package abstracts

import (
	"github.com/gorilla/securecookie"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/host/shttp"
)

type (
	OAuthClientHost struct {
		*BaseHost
		OAuthOptions        *OAuthOptions `json:"OAuth,omitempty"`
		HashKey             string
		BlockKey            string
		SignInPath          string
		SignInCallbackPath  string
		SignOutPath         string
		SignOutCallbackPath string
		OAuthClientHandler  IOAuthClientHandler
		ContextTokenStore   shttp.IContextTokenStore
		CookieProtoector    *securecookie.SecureCookie
	}
)

func (x *OAuthClientHost) BuildOAuthClientHost() {
	x.BaseHost.BuildBaseHost()

	if x.OAuthOptions == nil {
		log.Fatal("OAuth secion in configuration is missing")
	}
	x.OAuthOptions.BuildOAuthOptions()

	if x.BlockKey == "" {
		log.Fatal("block key cannot be empty")
	}
	if x.HashKey == "" {
		log.Fatal("hash key cannot be empty")
	}
	if x.SignInPath == "" {
		x.SignInPath = "/signin"
	}
	if x.SignInCallbackPath == "" {
		x.SignInCallbackPath = "/signin-oauth"
	}
	if x.SignOutPath == "" {
		x.SignOutPath = "/signout"
	}
	if x.SignOutCallbackPath == "" {
		x.SignOutCallbackPath = "/signout-oauth"
	}
}
