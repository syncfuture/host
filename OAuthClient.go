package host

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Lukiya/oauth2go"
	"github.com/Lukiya/oauth2go/core"
	oauth2core "github.com/Lukiya/oauth2go/core"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/securecookie"
	"github.com/kataras/iris/v12"
	iriscontext "github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/sessions"
	"github.com/syncfuture/go/config"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/srand"
	"github.com/syncfuture/go/u"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type (
	ClientUser struct {
		ID       string `json:"sub,omitempty"`
		Username string `json:"name,omitempty"`
		Email    string `json:"email,omitempty"`
		Role     int64  `json:"role,omitempty"`
		Level    int32  `json:"level,omitempty"`
		Status   int32  `json:"status,omitempty"`
	}
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

	OAuthClient struct {
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

func (t *ClientUser) UnmarshalJSON(d []byte) error {
	type T2 ClientUser // create new type with same structure as T but without its method set!
	x := struct {
		T2            // embed
		Role   string `json:"role,omitempty"`
		Level  string `json:"level,omitempty"`
		Status string `json:"status,omitempty"`
	}{T2: T2(*t)} // don't forget this, if you do and 't' already has some fields set you would lose them

	err := json.Unmarshal(d, &x)
	if u.LogError(err) {
		return err
	}

	*t = ClientUser(x.T2)
	var status, level int64
	t.Role, err = strconv.ParseInt(x.Role, 10, 64)
	u.LogError(err)
	status, err = strconv.ParseInt(x.Status, 10, 32)
	u.LogError(err)
	level, err = strconv.ParseInt(x.Level, 10, 32)
	u.LogError(err)

	t.Status = int32(status)
	t.Level = int32(level)
	return nil
}

func NewOAuthClientOptions(args ...string) *OAuthClientOptions {
	cp := config.NewJsonConfigProvider(args...)
	var options *OAuthClientOptions
	cp.GetStruct("OAuthClient", &options)
	if options == nil {
		log.Fatal("missing 'ClientServer' section in configuration")
	}
	options.ConfigProvider = cp
	return options
}

func NewOAuthClient(options *OAuthClientOptions) (r *OAuthClient) {
	// create pointer
	r = new(OAuthClient)
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
		Cookie:       "syncsession",
		Expires:      -1 * time.Hour,
		Encode:       r.CookieManager.Encode,
		Decode:       r.CookieManager.Decode,
		AllowReclaim: true,
	})
	r.SignInHandler = options.SignInHandler
	r.SignInCallbackHandler = options.SignInCallbackHandler
	r.SignOutHandler = options.SignOutHandler
	r.SignOutCallbackHandler = options.SignOutCallbackHandler

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
		log.Fatal("Cannot find 'ListenAddr' config")
	}
	x.IrisApp.Run(iris.Addr(x.ListenAddr))
}

func (x *OAuthClient) MvcAuthorize(ctx iriscontext.Context) {
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

func (x *OAuthClient) NewHttpClient(args ...interface{}) (*http.Client, error) {
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

func (x *OAuthClient) GetUser(ctx iriscontext.Context) (r *ClientUser) {
	session := x.SessionManager.Start(ctx)
	userJson := session.GetString(x.UserSessionName)
	if userJson != "" {
		// 已登录
		err := json.Unmarshal([]byte(userJson), &r)
		u.LogError(err)
	}
	return
}

func (x *OAuthClient) signinHanlder(ctx iriscontext.Context) {
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

func (x *OAuthClient) signInCallbackHandler(ctx iriscontext.Context) {
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

func (x *OAuthClient) signOutHandler(ctx iriscontext.Context) {
	session := x.SessionManager.Start(ctx)

	// 去Passport注销
	state := srand.String(32)
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

func (x *OAuthClient) signOutCallbackHandler(ctx iriscontext.Context) {
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

func (x *OAuthClient) saveToken(ctx iriscontext.Context, token *oauth2.Token) error {
	tokenJson, err := json.Marshal(token)
	if u.LogError(err) {
		return err
	}

	// 保存令牌到Session
	session := x.SessionManager.Start(ctx)
	session.Set(x.TokenSessionName, string(tokenJson))

	return nil
}

func (x *OAuthClient) getToken(ctx iriscontext.Context) (*oauth2.Token, error) {
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

func makeUserString(claims *jwt.MapClaims) string {
	bytes, err := json.Marshal(claims)
	if u.LogError(err) {
		return ""
	}
	return string(bytes)
}
