package host

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Lukiya/oauth2go"
	oauth2core "github.com/Lukiya/oauth2go/core"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/securecookie"
	"github.com/kataras/iris/v12"
	iriscontext "github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/sessions"
	"github.com/syncfuture/go/config"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/soidc"
	"github.com/syncfuture/go/srand"
	"github.com/syncfuture/go/u"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type (
	OAuthOptions struct {
		oauth2.Config
		PkceRequired     bool
		SignOutEndpoint  string
		ClientCredential *oauth2go.ClientCredential
	}
	ClientServerOptions struct {
		IrisBaseServerOptions
		AccessDeniedPath       string
		SignInPath             string
		SignInCallbackPath     string
		SignOutPath            string
		SignOutCallbackPath    string
		StaticFilesDir         string
		UserSessionName        string
		TokenSessionName       string
		HashKey                string
		BlockKey               string
		OAuth                  *OAuthOptions
		SignInHandler          iriscontext.Handler
		SignInCallbackHandler  iriscontext.Handler
		SignOutHandler         iriscontext.Handler
		SignOutCallbackHandler iriscontext.Handler
	}

	ClientServer struct {
		IrisBaseServer
		AccessDeniedPath       string
		SignInPath             string
		SignInCallbackPath     string
		SignOutPath            string
		SignOutCallbackPath    string
		StaticFilesDir         string
		UserSessionName        string
		TokenSessionName       string
		CookieManager          *securecookie.SecureCookie
		SessionManager         *sessions.Sessions
		OAuth                  *OAuthOptions
		SignInHandler          iriscontext.Handler
		SignInCallbackHandler  iriscontext.Handler
		SignOutHandler         iriscontext.Handler
		SignOutCallbackHandler iriscontext.Handler
	}
)

func NewClientServerOptions(args ...string) *ClientServerOptions {
	cp := config.NewJsonConfigProvider(args...)
	var options *ClientServerOptions
	cp.GetStruct("ClientServer", &options)
	if options == nil {
		log.Fatal("missing 'ClientServer' section in configuration")
	}
	options.ConfigProvider = cp
	return options
}

func NewClientServer(options *ClientServerOptions) (r *ClientServer) {
	// create pointer
	r = new(ClientServer)
	r.Name = options.Name
	r.configIrisBaseServer(&options.IrisBaseServerOptions)

	if options.OAuth == nil {
		log.Fatal("OAuth secion in configuration is missing")
	}
	if options.OAuth.Endpoint.AuthURL == "" {
		log.Fatal("OAuth.Endpoint.AuthURL cannot be empty")
	} else {
		options.OAuth.Endpoint.AuthURL = r.URLProvider.RenderURLCache(options.OAuth.Endpoint.AuthURL)
	}
	if options.OAuth.Endpoint.TokenURL == "" {
		log.Fatal("OAuth.Endpoint.TokenURL cannot be empty")
	} else {
		options.OAuth.Endpoint.TokenURL = r.URLProvider.RenderURLCache(options.OAuth.Endpoint.TokenURL)
	}
	if options.OAuth.RedirectURL == "" {
		log.Fatal("OAuth.RedirectURL cannot be empty")
	} else {
		options.OAuth.RedirectURL = r.URLProvider.RenderURLCache(options.OAuth.RedirectURL)
	}
	if options.OAuth.SignOutEndpoint == "" {
		log.Fatal("OAuthSignOutEndpoint cannot be empty")
	} else {
		options.OAuth.SignOutEndpoint = r.URLProvider.RenderURLCache(options.OAuth.SignOutEndpoint)
	}
	options.OAuth.ClientCredential = &oauth2go.ClientCredential{
		Config: &clientcredentials.Config{
			ClientID:     options.OAuth.ClientID,
			ClientSecret: options.OAuth.ClientSecret,
			TokenURL:     options.OAuth.Endpoint.TokenURL,
			Scopes:       options.OAuth.Scopes,
		},
	}

	if options.AccessDeniedPath == "" {
		options.AccessDeniedPath = "/accessdenied"
	}
	if options.SignInPath == "" {
		options.SignInPath = "/signin"
	}
	if options.SignInCallbackPath == "" {
		options.SignInCallbackPath = "/signin-oauth"
	}
	if options.SignOutPath == "" {
		options.SignOutPath = "/signout"
	}
	if options.SignOutCallbackPath == "" {
		options.SignOutCallbackPath = "/signout-oauth"
	}
	if options.UserSessionName == "" {
		options.UserSessionName = "OAuth.u001"
	}
	if options.TokenSessionName == "" {
		options.TokenSessionName = "OAuth.t001"
	}
	if options.StaticFilesDir == "" {
		options.StaticFilesDir = "./wwwroot"
	}
	if options.SignInHandler == nil {
		options.SignInHandler = r.signinHanlder
	}
	if options.SignInCallbackHandler == nil {
		options.SignInCallbackHandler = r.signInCallbackHandler
	}
	if options.SignOutHandler == nil {
		options.SignOutHandler = r.signOutHandler
	}
	if options.SignOutCallbackHandler == nil {
		options.SignOutCallbackHandler = r.signOutCallbackHandler
	}

	r.OAuth = options.OAuth
	// r.OAuthSignOutEndpoint = options.OAuthSignOutEndpoint
	r.SignInPath = options.SignInPath
	r.SignInCallbackPath = options.SignInCallbackPath
	r.SignOutPath = options.SignOutPath
	r.SignOutCallbackPath = options.SignOutCallbackPath
	r.AccessDeniedPath = options.AccessDeniedPath
	r.StaticFilesDir = options.StaticFilesDir
	r.UserSessionName = options.UserSessionName
	r.TokenSessionName = options.TokenSessionName
	r.CookieManager = securecookie.New([]byte(options.HashKey), []byte(options.BlockKey))
	r.SessionManager = sessions.New(sessions.Config{
		Expires:      -1 * time.Hour,
		Cookie:       soidc.COKI_SESSION,
		Encode:       r.CookieManager.Encode,
		Decode:       r.CookieManager.Decode,
		AllowReclaim: true,
	})
	r.SignInHandler = options.SignInHandler
	r.SignInCallbackHandler = options.SignInCallbackHandler
	r.SignInHandler = options.SignInHandler
	r.SignInHandler = options.SignInHandler

	// 添加内置终结点
	r.IrisApp.Get(r.SignInPath, r.SignInHandler)
	r.IrisApp.Get(r.SignInCallbackPath, r.SignInCallbackHandler)
	r.IrisApp.Get(r.SignOutPath, r.SignOutHandler)
	r.IrisApp.Get(r.SignOutCallbackPath, r.SignOutCallbackHandler)

	// 注册视图引擎
	if r.ViewEngine == nil {
		r.ViewEngine = iris.HTML("./views", ".html").Layout("shared/_layout.html").Reload(r.Debug)
	}
	r.IrisApp.RegisterView(r.ViewEngine)

	// 注册静态文件
	r.IrisApp.HandleDir("/", r.StaticFilesDir)

	return
}

func (x *ClientServer) Run(actionGroups ...*[]*Action) {
	// 构造页面路由字典
	actionMap := make(map[string]*Action)
	for _, actionGroup := range actionGroups {
		for _, action := range *actionGroup {
			actionMap[action.Route] = action
		}
	}
	x.ActionMap = &actionMap

	x.registerActions()

	if x.ListenAddr == "" {
		log.Fatal("Cannot find 'ListenAddr' config")
	}
	x.IrisApp.Run(iris.Addr(x.ListenAddr))
}

func (x *ClientServer) Authorize(ctx iriscontext.Context) {
	session := x.SessionManager.Start(ctx)

	// handlerName := ctx.GetCurrentRoute().MainHandlerName()
	route := ctx.GetCurrentRoute().Name()
	var area, controller, action string
	if act, ok := (*x.ActionMap)[route]; ok {
		area = act.Area
		controller = act.Controller
		action = act.Action
	}
	// area, controller, action := getRoutes(route)

	// 判断请求是否允许访问
	userStr := session.GetString(x.UserSessionName)
	if userStr != "" {
		userArray := strings.Split(userStr, "|")
		if len(userArray) == 6 {
			// userId := userArray[0]
			roles, err := strconv.ParseInt(userArray[2], 10, 64)
			if u.LogError(err) {
				return
			}
			level, err := strconv.Atoi(userArray[3])
			if u.LogError(err) {
				return
			}
			// 已登录
			allow := x.PermissionAuditor.CheckRouteWithLevel(area, controller, action, roles, int32(level))
			if allow {
				// 有权限
				ctx.Next()
				return
			} else {
				// 没权限
				ctx.Redirect(x.AccessDeniedPath, http.StatusFound)
				return
			}
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
	session.Set(state, ctx.Request().URL.String())
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

func (x *ClientServer) signinHanlder(ctx iriscontext.Context) {
	returnURL := ctx.FormValue(oauth2core.Form_ReturnUrl)
	if returnURL == "" {
		returnURL = "/"
	}

	session := x.SessionManager.Start(ctx)
	userStr := session.GetString(x.UserSessionName)
	if userStr != "" {
		// 已登录
		ctx.Redirect(returnURL, http.StatusFound)
		return
	}

	// 记录请求地址，跳转去登录页面
	state := srand.String(32)
	session.Set(state, returnURL)
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

func (x *ClientServer) signInCallbackHandler(ctx iriscontext.Context) {
	session := x.SessionManager.Start(ctx)

	state := ctx.FormValue(oauth2core.Form_State)
	redirectUrl := session.GetString(state)
	if redirectUrl == "" {
		ctx.WriteString("invalid state")
		ctx.StatusCode(http.StatusBadRequest)
		return
	}
	session.Delete(state) // 释放内存

	var sessionCodeVerifier, sessionSodeChallengeMethod string
	if x.OAuth.PkceRequired {
		sessionCodeVerifier = session.GetString(oauth2core.Form_CodeVerifier)
		if sessionCodeVerifier == "" {
			ctx.WriteString("pkce code verifier does not exist in store")
			ctx.StatusCode(http.StatusBadRequest)
			return
		}
		session.Delete(oauth2core.Form_CodeVerifier)
		sessionSodeChallengeMethod = session.GetString(oauth2core.Form_CodeChallengeMethod)
		if sessionCodeVerifier == "" {
			ctx.WriteString("pkce transformation method does not exist in store")
			ctx.StatusCode(http.StatusBadRequest)
			return
		}
		session.Delete(oauth2core.Form_CodeChallengeMethod)

		codeChallenge := ctx.FormValue(oauth2core.Form_CodeChallenge)
		codeChallengeMethod := ctx.FormValue(oauth2core.Form_CodeChallengeMethod)

		if sessionSodeChallengeMethod != codeChallengeMethod {
			ctx.WriteString("pkce transformation method does not match")
			log.Debugf("session method: '%s', incoming method:'%s'", sessionSodeChallengeMethod, codeChallengeMethod)
			ctx.StatusCode(http.StatusBadRequest)
			return
		} else if (sessionSodeChallengeMethod == oauth2core.Pkce_Plain && codeChallenge != oauth2core.ToSHA256Base64URL(sessionCodeVerifier)) ||
			(sessionSodeChallengeMethod == oauth2core.Pkce_Plain && codeChallenge != sessionCodeVerifier) {
			ctx.WriteString("pkce code verifiver and chanllenge does not match")
			log.Debugf("session verifiver: '%s', incoming chanllenge:'%s'", sessionCodeVerifier, codeChallenge)
			ctx.StatusCode(http.StatusBadRequest)
			return
		}
	}

	// 交换令牌
	code := ctx.FormValue(oauth2core.Form_Code)
	httpCtx := context.Background()
	var oauth2Token *oauth2.Token
	var err error

	if x.OAuth.PkceRequired {
		codeChanllengeParam := oauth2.SetAuthURLParam(oauth2core.Form_CodeVerifier, sessionCodeVerifier)
		codeChanllengeMethodParam := oauth2.SetAuthURLParam(oauth2core.Form_CodeChallengeMethod, sessionSodeChallengeMethod)
		oauth2Token, err = x.OAuth.Exchange(httpCtx, code, codeChanllengeParam, codeChanllengeMethodParam)
	} else {
		oauth2Token, err = x.OAuth.Exchange(httpCtx, code)
	}

	if u.LogError(err) {
		ctx.WriteString(err.Error())
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}

	// 将字符串转化为令牌对象，忽略KeyFunc不存在错误
	jwtToken, err := new(jwt.Parser).Parse(oauth2Token.AccessToken, nil)
	vErr := err.(*jwt.ValidationError)
	if vErr.Errors != jwt.ValidationErrorUnverifiable {
		ctx.WriteString(err.Error())
		u.LogError(err)
		return
	}
	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if ok {
		userStr := makeUserString(&claims)
		session.Set(x.UserSessionName, userStr)

		// // 保存令牌
		// x.SaveToken(ctx, oauth2Token)

		// 重定向到登录前页面
		ctx.Redirect(redirectUrl, http.StatusFound)
	} else {
		err = errors.New("cannot convert jwtToken.Claims to jwt.MapClaims")
		ctx.WriteString(err.Error())
		u.LogError(err)
	}
}

func (x *ClientServer) signOutHandler(ctx iriscontext.Context) {
	session := x.SessionManager.Start(ctx)

	// 去Passport注销
	state := srand.String(32)
	session.Set(state, ctx.FormValue(oauth2core.Form_ReturnUrl))
	// signoutUrl := x.OAuthSignOutEndpoint + "?post_logout_redirect_uri=" + url.PathEscape(x.SignInCallbackPath) + "&id_token_hint=" + idToken + "&state=" + state
	signoutUrl := x.OAuth.SignOutEndpoint + "?post_logout_redirect_uri=" + url.PathEscape(x.SignInCallbackPath) + "&state=" + state
	ctx.Redirect(signoutUrl, http.StatusFound)
}

func (x *ClientServer) signOutCallbackHandler(ctx iriscontext.Context) {
	session := x.SessionManager.Start(ctx)

	state := ctx.FormValue(oauth2core.Form_State)
	redirectUrl := session.GetString(state)
	if redirectUrl == "" {
		ctx.WriteString("invalid state")
		ctx.StatusCode(http.StatusBadRequest)
		return
	}
	session.Delete(state)
	session.Delete(x.UserSessionName)
	session.Delete(x.TokenSessionName)
	session.Destroy()

	// 跳转回登出时的页面
	ctx.Redirect(redirectUrl, http.StatusFound)
}

func (x *ClientServer) saveToken(ctx iriscontext.Context, token *oauth2.Token) error {
	tokenJson, err := json.Marshal(token)
	if u.LogError(err) {
		return err
	}

	// 保存令牌到Session
	session := x.SessionManager.Start(ctx)
	session.Set(x.TokenSessionName, string(tokenJson))

	return nil
}

func (x *ClientServer) getToken(ctx iriscontext.Context) (*oauth2.Token, error) {
	// 从Session获取令牌
	session := x.SessionManager.Start(ctx)
	tokenJson := session.GetString(x.TokenSessionName)

	t := new(oauth2.Token)
	err := json.Unmarshal([]byte(tokenJson), t)
	if u.LogError(err) {
		return nil, err
	}

	return t, err
}

func (x *ClientServer) NewHttpClient(args ...interface{}) (*http.Client, error) {
	goctx := context.Background()
	if len(args) == 0 {
		return x.OAuth.ClientCredential.Client(goctx), nil
	}

	irisctx, ok := args[0].(iriscontext.Context)
	if !ok {
		panic("first parameter must be iris context")
	}

	session := x.SessionManager.Start(irisctx)
	userStr := session.GetString(x.UserSessionName)
	if userStr == "" {
		return http.DefaultClient, fmt.Errorf("user doesn't login")
	}

	t, err := x.getToken(irisctx)
	if u.LogError(err) {
		return http.DefaultClient, err
	}

	tokenSource := x.OAuth.TokenSource(goctx, t)
	newToken, err := tokenSource.Token()
	if u.LogError(err) {
		return http.DefaultClient, err
	}

	if newToken.AccessToken != t.AccessToken {
		x.saveToken(irisctx, newToken)
	}

	return oauth2.NewClient(goctx, tokenSource), nil
}

func makeUserString(claims *jwt.MapClaims) string {
	sub := getClaimString(claims, "sub")
	name := getClaimString(claims, "name")
	role := getClaimString(claims, "role")
	level := getClaimString(claims, "level")
	status := getClaimString(claims, "status")
	email := getClaimString(claims, "email")

	return fmt.Sprintf("%s|%s|%s|%s|%s|%s", sub, name, role, level, status, email)
}

func getClaimString(claims *jwt.MapClaims, name string) string {
	v := (*claims)[name]
	if v == nil {
		return ""
	}
	return v.(string)
}
