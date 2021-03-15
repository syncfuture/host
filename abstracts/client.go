package abstracts

import (
	"net/http"

	oauth2go "github.com/Lukiya/oauth2go/core"
	"github.com/gorilla/securecookie"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/srand"
	"github.com/syncfuture/go/ssecurity"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host/shttp"
	"golang.org/x/oauth2"
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

type ClientAuthMiddleware struct {
	UserJsonSessionkey string
	AccessDeniedPath   string
	OAuthOptions       *OAuthOptions
	PermissionAuditor  ssecurity.IPermissionAuditor
}

func newClientAuthMiddleware(
	userJsonSessionkey string,
	accessDeniedPath string,
	oauthOptions *OAuthOptions,
	permissionAuditor ssecurity.IPermissionAuditor,
) IAuthMiddleware {
	return &ClientAuthMiddleware{
		UserJsonSessionkey: userJsonSessionkey,
		AccessDeniedPath:   accessDeniedPath,
		OAuthOptions:       oauthOptions,
		PermissionAuditor:  permissionAuditor,
	}
}

func (x *ClientAuthMiddleware) Serve(next shttp.RequestHandler, routes ...string) shttp.RequestHandler {
	var area, controller, action string
	count := len(routes)
	if count == 0 || count > 3 {
		log.Fatal("invalid routes array")
	}

	area = routes[0]
	if count >= 2 {
		controller = routes[1]
	}
	if count == 3 {
		action = routes[2]
	}

	return func(ctx shttp.IHttpContext) {
		user := shttp.GetUser(ctx, x.UserJsonSessionkey)

		// 判断请求是否允许访问
		if user != nil {
			if x.PermissionAuditor.CheckRouteWithLevel(area, controller, action, user.Role, user.Level) {
				// 有权限
				next(ctx)
				return
			} else {
				// 没权限
				ctx.Redirect(x.AccessDeniedPath, http.StatusFound)
				return
			}
		}

		// 未登录
		allow := x.PermissionAuditor.CheckRouteWithLevel(area, controller, action, 0, 0)
		if allow {
			// 允许匿名
			next(ctx)
			return
		}

		// 记录请求地址，跳转去登录页面
		state := srand.String(32)
		ctx.SetSession(state, ctx.RequestURL())
		if x.OAuthOptions.PkceRequired {
			codeVerifier := oauth2go.Random64String()
			codeChanllenge := oauth2go.ToSHA256Base64URL(codeVerifier)
			ctx.SetSession(oauth2go.Form_CodeVerifier, codeVerifier)
			ctx.SetSession(oauth2go.Form_CodeChallengeMethod, oauth2go.Pkce_S256)
			codeChanllengeParam := oauth2.SetAuthURLParam(oauth2go.Form_CodeChallenge, codeChanllenge)
			codeChanllengeMethodParam := oauth2.SetAuthURLParam(oauth2go.Form_CodeChallengeMethod, oauth2go.Pkce_S256)
			ctx.Redirect(x.OAuthOptions.AuthCodeURL(state, codeChanllengeParam, codeChanllengeMethodParam), http.StatusFound)
		} else {
			ctx.Redirect(x.OAuthOptions.AuthCodeURL(state), http.StatusFound)
		}
	}
}
