package host

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/Lukiya/oauth2go"
	"github.com/Lukiya/oauth2go/core"
	oauth2core "github.com/Lukiya/oauth2go/core"
	"github.com/gorilla/securecookie"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
	"github.com/muesli/cache2go"
	"github.com/pascaldekloe/jwt"
	config "github.com/syncfuture/go/sconfig"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/srand"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host/model"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const _cookieTokenProtectorKey = "token"

type (
	OAuthOptions struct {
		oauth2.Config
		PkceRequired       bool
		EndSessionEndpoint string
		SignOutRedirectURL string
		ClientCredential   *oauth2go.ClientCredential
	}
	OAuthClientOptions struct {
		IrisBaseServerOptions
		AccessDeniedPath       string
		SignInPath             string
		SignInCallbackPath     string
		SignOutPath            string
		SignOutCallbackPath    string
		StaticFilesDir         string
		LayoutTemplate         string
		ViewsExtension         string
		SessionName            string
		TokenCookieName        string
		HashKey                string
		BlockKey               string
		OAuth                  *OAuthOptions
		SignInHandler          iris.Handler
		SignInCallbackHandler  iris.Handler
		SignOutHandler         iris.Handler
		SignOutCallbackHandler iris.Handler
	}

	OAuthClient struct {
		IrisBaseServer
		AccessDeniedPath       string
		SignInPath             string
		SignInCallbackPath     string
		SignOutPath            string
		SignOutCallbackPath    string
		StaticFilesDir         string
		ViewsExtension         string
		LayoutTemplate         string
		SessionName            string
		TokenCookieName        string
		userJsonSessionkey     string
		userIDSessionKey       string
		CookieProtoector       *securecookie.SecureCookie
		SessionManager         *sessions.Sessions
		UserLocks              *cache2go.CacheTable
		OAuth                  *OAuthOptions
		SignInHandler          iris.Handler
		SignInCallbackHandler  iris.Handler
		SignOutHandler         iris.Handler
		SignOutCallbackHandler iris.Handler
	}
)

func NewOAuthClientOptions(args ...string) *OAuthClientOptions {
	cp := config.NewJsonConfigProvider(args...)
	var options *OAuthClientOptions
	cp.GetStruct("OAuthClient", &options)
	if options == nil {
		log.Fatal("missing 'OAuthClient' section in configuration")
	}
	options.ConfigProvider = cp
	return options
}

func NewOAuthClient(options *OAuthClientOptions) (r *OAuthClient) {
	// create pointer
	r = new(OAuthClient)
	r.configIrisBaseServer(&options.IrisBaseServerOptions)
	// r.Name = options.Name
	// r.URIKey = options.URIKey
	// r.RouteKey = options.RouteKey
	// r.PermissionKey = options.PermissionKey

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
	if options.OAuth.SignOutRedirectURL == "" {
		log.Fatal("OAuth.SignOutRedirectURL cannot be empty")
	} else {
		options.OAuth.SignOutRedirectURL = r.URLProvider.RenderURLCache(options.OAuth.SignOutRedirectURL)
	}
	if options.OAuth.EndSessionEndpoint == "" {
		log.Fatal("OAuth.EndSessionEndpoint cannot be empty")
	} else {
		options.OAuth.EndSessionEndpoint = r.URLProvider.RenderURLCache(options.OAuth.EndSessionEndpoint)
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
	if options.SessionName == "" {
		options.SessionName = "go.cookie1"
	}
	if options.TokenCookieName == "" {
		options.TokenCookieName = "go.cookie2"
	}
	if options.StaticFilesDir == "" {
		options.StaticFilesDir = "./wwwroot"
	}
	if options.ViewsExtension == "" {
		options.ViewsExtension = ".html"
	}
	if options.LayoutTemplate == "" {
		options.LayoutTemplate = "shared/_layout.html"
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
	r.ViewsDir = options.ViewsDir
	r.LayoutTemplate = options.LayoutTemplate
	r.ViewsExtension = options.ViewsExtension
	r.SessionName = options.SessionName
	r.TokenCookieName = options.TokenCookieName
	r.userIDSessionKey = "UserID"
	r.userJsonSessionkey = "UserJson"
	r.CookieProtoector = securecookie.New([]byte(options.HashKey), []byte(options.BlockKey))
	r.SessionManager = sessions.New(sessions.Config{
		Cookie:                      r.SessionName,
		Expires:                     time.Duration(-1),
		Encoding:                    r.CookieProtoector,
		AllowReclaim:                true,
		DisableSubdomainPersistence: true,
	})
	r.SignInHandler = options.SignInHandler
	r.SignInCallbackHandler = options.SignInCallbackHandler
	r.SignOutHandler = options.SignOutHandler
	r.SignOutCallbackHandler = options.SignOutCallbackHandler
	r.UserLocks = cache2go.Cache("UserLocks")

	// 添加内置终结点
	r.IrisApp.Get(r.SignInPath, r.SignInHandler)
	r.IrisApp.Get(r.SignInCallbackPath, r.SignInCallbackHandler)
	r.IrisApp.Get(r.SignOutPath, r.SignOutHandler)
	r.IrisApp.Get(r.SignOutCallbackPath, r.SignOutCallbackHandler)

	// 注册视图引擎
	if r.ViewEngine == nil {
		r.ViewEngine = iris.HTML(r.ViewsDir, r.ViewsExtension).Layout(r.LayoutTemplate).Reload(r.Debug)
	}
	r.IrisApp.RegisterView(r.ViewEngine)

	// 注册静态文件
	r.IrisApp.HandleDir("/", r.StaticFilesDir)

	return
}

func (x *OAuthClient) Run(actionGroups ...*[]*Action) {
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
		log.Fatal("cannot find 'ListenAddr' in config")
	}
	x.IrisApp.Run(iris.Addr(x.ListenAddr))
}

func (x *OAuthClient) MvcAuthorize(ctx iris.Context) {
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
	user := x.GetUser(ctx)
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

func (x *OAuthClient) Client() (*http.Client, error) {
	return x.OAuth.ClientCredential.Client(context.Background()), nil
}

func (x *OAuthClient) GetUserLock(userID string) *sync.RWMutex {
	if !x.UserLocks.Exists(userID) {
		x.UserLocks.Add(userID, time.Second*30, new(sync.RWMutex))
	}

	userLockCache, _ := x.UserLocks.Value(userID)
	return userLockCache.Data().(*sync.RWMutex)
}

func (x *OAuthClient) UserClient(ctx iris.Context) (*http.Client, error) {
	goctx := context.Background()
	userID := x.GetUserID(ctx)
	if userID == "" {
		return http.DefaultClient, nil
	}

	// 获取用户锁
	userLock := x.GetUserLock(userID)

	// read lock
	userLock.RLock()
	t, err := x.getToken(ctx)
	userLock.RUnlock()

	if err != nil {
		return http.DefaultClient, err
	}

	tokenSource := x.OAuth.TokenSource(goctx, t)
	newToken, err := tokenSource.Token()
	if err != nil {
		// refresh token failed, sign user out
		x.SignOut(ctx)
		return http.DefaultClient, err
	}

	if newToken.AccessToken != t.AccessToken {
		// token been refreshed, lock
		userLock.Lock()
		// save token to session
		x.saveToken(ctx, newToken)
		// unlock
		defer userLock.Unlock()
	}

	return oauth2.NewClient(goctx, tokenSource), nil
}

func (x *OAuthClient) GetUser(ctx iris.Context) (r *model.User) {
	session := x.SessionManager.Start(ctx)
	userJson := session.GetString(x.userJsonSessionkey)
	if userJson != "" {
		// 已登录
		err := json.Unmarshal([]byte(userJson), &r)
		u.LogError(err)
	}
	return
}

func (x *OAuthClient) GetUserID(ctx iris.Context) string {
	session := x.SessionManager.Start(ctx)
	return session.GetString(x.userIDSessionKey)
}

func (x *OAuthClient) signinHanlder(ctx iris.Context) {
	returnURL := ctx.FormValue(oauth2core.Form_ReturnUrl)
	if returnURL == "" {
		returnURL = "/"
	}

	session := x.SessionManager.Start(ctx)
	userStr := session.GetString(x.userJsonSessionkey)
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

func (x *OAuthClient) signInCallbackHandler(ctx iris.Context) {
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

	// 获取老的刷新令牌，发送给Auth服务器进行注销
	token, _ := x.getToken(ctx)
	var refreshTokenOption oauth2.AuthCodeOption
	if token != nil && token.RefreshToken != "" {
		refreshTokenOption = oauth2.SetAuthURLParam(oauth2core.Form_RefreshToken, token.RefreshToken)
	}

	if x.OAuth.PkceRequired {
		codeChanllengeParam := oauth2.SetAuthURLParam(oauth2core.Form_CodeVerifier, sessionCodeVerifier)
		codeChanllengeMethodParam := oauth2.SetAuthURLParam(oauth2core.Form_CodeChallengeMethod, sessionSodeChallengeMethod)

		// 发送交换令牌请求
		if refreshTokenOption != nil {
			oauth2Token, err = x.OAuth.Exchange(httpCtx, code, codeChanllengeParam, codeChanllengeMethodParam, refreshTokenOption)
		} else {
			oauth2Token, err = x.OAuth.Exchange(httpCtx, code, codeChanllengeParam, codeChanllengeMethodParam)
		}
	} else {
		if refreshTokenOption != nil {
			oauth2Token, err = x.OAuth.Exchange(httpCtx, code, refreshTokenOption)
		} else {
			oauth2Token, err = x.OAuth.Exchange(httpCtx, code)
		}
	}

	if u.LogError(err) {
		ctx.WriteString(err.Error())
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}

	// 将字符串转化为令牌对象
	jwtToken, err := jwt.ParseWithoutCheck([]byte(oauth2Token.AccessToken))
	if err == nil {
		userStr := makeUserString(jwtToken)
		session.Set(x.userJsonSessionkey, userStr)
		if jwtToken.Subject != "" {
			session.Set(x.userIDSessionKey, jwtToken.Subject)
		}

		// 保存令牌
		x.saveToken(ctx, oauth2Token)

		// 重定向到登录前页面
		ctx.Redirect(redirectUrl, http.StatusFound)
	} else {
		ctx.WriteString(err.Error())
		u.LogError(err)
	}
}

func (x *OAuthClient) signOutHandler(ctx iris.Context) {
	session := x.SessionManager.Start(ctx)

	// 去Passport注销
	state := srand.String(32)
	returnUrl := ctx.FormValue(oauth2core.Form_ReturnUrl)
	if returnUrl == "" {
		returnUrl = "/"
	}
	session.Set(state, ctx.FormValue(oauth2core.Form_ReturnUrl))
	targetURL := fmt.Sprintf("%s?%s=%s&%s=%s&%s=%s",
		x.OAuth.EndSessionEndpoint,
		core.Form_ClientID,
		x.OAuth.ClientID,
		core.Form_RedirectUri,
		url.PathEscape(x.OAuth.SignOutRedirectURL),
		core.Form_State,
		url.QueryEscape(state),
	)
	ctx.Redirect(targetURL, http.StatusFound)
}

func (x *OAuthClient) signOutCallbackHandler(ctx iris.Context) {
	session := x.SessionManager.Start(ctx)

	state := ctx.FormValue(oauth2core.Form_State)
	returnURL := session.GetString(state)
	if returnURL == "" {
		ctx.WriteString("invalid state")
		ctx.StatusCode(http.StatusBadRequest)
		return
	}

	endSessionID := ctx.FormValue(oauth2core.Form_EndSessionID)
	if endSessionID == "" {
		ctx.WriteString("missing es_id")
		ctx.StatusCode(http.StatusBadRequest)
		return
	}

	token, _ := x.getToken(ctx)
	if token != nil {
		// 请求Auth服务器删除老RefreshToken
		data := make(url.Values, 5)
		data[oauth2core.Form_State] = []string{state}
		data[oauth2core.Form_EndSessionID] = []string{endSessionID}
		data[oauth2core.Form_ClientID] = []string{x.OAuth.ClientID}
		data[oauth2core.Form_ClientSecret] = []string{x.OAuth.ClientSecret}
		data[oauth2core.Form_RefreshToken] = []string{token.RefreshToken}
		http.PostForm(x.OAuth.EndSessionEndpoint, data)
	}

	x.SignOut(ctx)
	// 跳转回登出时的页面
	ctx.Redirect(returnURL, http.StatusFound)
}

func (x *OAuthClient) SignOut(ctx iris.Context) {
	x.SessionManager.Destroy(ctx)
	ctx.RemoveCookie(x.TokenCookieName)
}

/// saveToken 保存令牌
func (x *OAuthClient) saveToken(ctx iris.Context, token *oauth2.Token) error {
	tokenJson, err := json.Marshal(token)
	if err != nil {
		return err
	}

	// 令牌加密
	securedString, err := x.CookieProtoector.Encode(_cookieTokenProtectorKey, tokenJson)

	// 保存加密后的令牌到Cookie
	tokenCookie := new(http.Cookie)
	tokenCookie.Name = x.TokenCookieName
	tokenCookie.Value = securedString
	tokenCookie.HttpOnly = true
	tokenClaims, err := jwt.ParseWithoutCheck([]byte(token.AccessToken))
	if err != nil {
		return err
	}
	if rexp, ok := tokenClaims.Set[oauth2core.Claim_RefreshTokenExpire].(float64); ok {
		// claims里有刷新令牌过期时间，作为Cookie
		tokenCookie.Expires = time.Unix(int64(rexp), 0)
		// 否则作为session存储
	}

	ctx.SetCookie(tokenCookie)
	return nil
}

/// getToken 获取令牌
func (x *OAuthClient) getToken(ctx iris.Context) (*oauth2.Token, error) {
	// 从Session获取令牌
	tokenJson := ctx.GetCookie(x.TokenCookieName)
	if tokenJson == "" {
		return nil, nil
	}
	var tokenJsonBytes []byte
	err := x.CookieProtoector.Decode(_cookieTokenProtectorKey, tokenJson, &tokenJsonBytes)
	if err != nil {
		return nil, err
	}

	t := new(oauth2.Token)
	err = json.Unmarshal(tokenJsonBytes, t)
	if err != nil {
		return nil, err
	}

	return t, err
}

func makeUserString(token *jwt.Claims) (r string) {
	r = string(token.Raw)
	return
}
