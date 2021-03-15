package abstracts

import (
	"net/http"

	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/ssecurity"
	"github.com/syncfuture/host/shttp"
)

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
		redirectAuthorizeEndpoint(ctx, x.OAuthOptions, ctx.RequestURL())

		// state := srand.String(32)
		// ctx.SetSession(state, ctx.RequestURL())
		// if x.OAuthOptions.PkceRequired {
		// 	codeVerifier := oauth2core.Random64String()
		// 	codeChanllenge := oauth2core.ToSHA256Base64URL(codeVerifier)
		// 	ctx.SetSession(oauth2core.Form_CodeVerifier, codeVerifier)
		// 	ctx.SetSession(oauth2core.Form_CodeChallengeMethod, oauth2core.Pkce_S256)
		// 	codeChanllengeParam := oauth2.SetAuthURLParam(oauth2core.Form_CodeChallenge, codeChanllenge)
		// 	codeChanllengeMethodParam := oauth2.SetAuthURLParam(oauth2core.Form_CodeChallengeMethod, oauth2core.Pkce_S256)
		// 	ctx.Redirect(x.OAuthOptions.AuthCodeURL(state, codeChanllengeParam, codeChanllengeMethodParam), http.StatusFound)
		// } else {
		// 	ctx.Redirect(x.OAuthOptions.AuthCodeURL(state), http.StatusFound)
		// }
	}
}
