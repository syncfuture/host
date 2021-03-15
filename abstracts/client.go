package abstracts

import (
	"github.com/gorilla/securecookie"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host/shttp"
)

type (
	OAuthClientHost struct {
		*BaseHost
		OAuthOptions        *OAuthOptions `json:"OAuth,omitempty"`
		HashKey             string
		BlockKey            string
		UserJsonSessionKey  string
		UserIDSessionKey    string
		TokenCookieName     string
		SignInPath          string
		SignInCallbackPath  string
		SignOutPath         string
		SignOutCallbackPath string
		AccessDeniedPath    string
		OAuthClientHandler  IOAuthClientHandler
		AuthMiddleware      IAuthMiddleware
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
	if x.AccessDeniedPath == "" {
		x.AccessDeniedPath = "/accessdenied"
	}
	if x.UserJsonSessionKey == "" {
		x.UserJsonSessionKey = "USERJSON"
	}
	if x.UserIDSessionKey == "" {
		x.UserIDSessionKey = "USERID"
	}
	if x.TokenCookieName == "" {
		x.TokenCookieName = "go.cookie2"
	}

	////////// cookie protoector
	if x.CookieProtoector == nil {
		x.CookieProtoector = securecookie.New(u.StrToBytes(x.HashKey), u.StrToBytes(x.BlockKey))
	}

	////////// context token store
	if x.ContextTokenStore == nil {
		x.ContextTokenStore = shttp.NewCookieTokenStore(x.TokenCookieName, x.CookieProtoector)
	}

	////////// oauth client handler
	if x.OAuthClientHandler == nil {
		x.OAuthClientHandler = NewDefaultOAuthClientHandler(x.OAuthOptions, x.ContextTokenStore, x.UserJsonSessionKey, x.UserIDSessionKey, x.TokenCookieName)
	}

	////////// auth middleware
	if x.AuthMiddleware == nil {
		x.AuthMiddleware = newClientAuthMiddleware(x.UserJsonSessionKey, x.AccessDeniedPath, x.OAuthOptions, x.PermissionAuditor)
	}
}
