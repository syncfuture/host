package siris

import (
	"encoding/json"
	"net/http"
	"strings"

	oauth2core "github.com/Lukiya/oauth2go/core"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
	"github.com/syncfuture/go/srand"
	"github.com/syncfuture/go/ssecurity"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host/abstracts"
	"github.com/syncfuture/host/model"
	"golang.org/x/oauth2"
)

type AuthMiddleware struct {
	PermissionAuditor  ssecurity.IPermissionAuditor
	SessionManager     *sessions.Sessions
	OAuth              *abstracts.OAuthOptions
	AccessDeniedPath   string
	userJsonSessionkey string
}

func NewAuthMiddleware(
	permissionAuditor ssecurity.IPermissionAuditor,
	sessionManager *sessions.Sessions,
	oauthOptions *abstracts.OAuthOptions,
	accessDeniedPath string,
	userJsonSessionkey string,
) *AuthMiddleware {
	return &AuthMiddleware{
		PermissionAuditor:  permissionAuditor,
		SessionManager:     sessionManager,
		OAuth:              oauthOptions,
		AccessDeniedPath:   accessDeniedPath,
		userJsonSessionkey: userJsonSessionkey,
	}
}

func (x *AuthMiddleware) Serve(ctx iris.Context) {
	session := x.SessionManager.Start(ctx)

	handlerName := ctx.GetCurrentRoute().MainHandlerName()
	area, controller, action := getRoutes(handlerName)
	// route := ctx.GetCurrentRoute().Name()
	// var area, controller, action string
	// if act, ok := (*x.ActionMap)[route]; ok {
	// 	area = act.Area
	// 	controller = act.Controller
	// 	action = act.Action
	// }

	// 判断请求是否允许访问
	user := x.getUser(ctx)
	if user != nil {
		if x.PermissionAuditor.CheckRouteWithLevel(area, controller, action, user.Role, user.Level) {
			// 有权限
			ctx.Next()
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
		ctx.Next()
		return
	}

	// 记录请求地址，跳转去登录页面
	state := srand.String(32)
	session.Set(state, ctx.Request().RequestURI)
	if x.OAuth.PkceRequired {
		codeVerifier := oauth2core.Random64String()
		codeChanllenge := oauth2core.ToSHA256Base64URL(codeVerifier)
		session.Set(oauth2core.Form_CodeVerifier, codeVerifier)
		session.Set(oauth2core.Form_CodeChallengeMethod, oauth2core.Pkce_S256)
		codeChanllengeParam := oauth2.SetAuthURLParam(oauth2core.Form_CodeChallenge, codeChanllenge)
		codeChanllengeMethodParam := oauth2.SetAuthURLParam(oauth2core.Form_CodeChallengeMethod, oauth2core.Pkce_S256)
		ctx.Redirect(x.OAuth.AuthCodeURL(state, codeChanllengeParam, codeChanllengeMethodParam), http.StatusFound)
	} else {
		ctx.Redirect(x.OAuth.AuthCodeURL(state), http.StatusFound)
	}
}

func (x *AuthMiddleware) getUser(ctx iris.Context) (r *model.User) {
	session := x.SessionManager.Start(ctx)
	userJson := session.GetString(x.userJsonSessionkey)
	if userJson != "" {
		// 已登录
		err := json.Unmarshal([]byte(userJson), &r)
		u.LogError(err)
	}
	return
}

func getRoutes(handlerName string) (string, string, string) {
	array := strings.Split(handlerName, ".")
	return array[0], array[1], array[2]
}
