package siris

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/Lukiya/oauth2go"
	oauth2core "github.com/Lukiya/oauth2go/core"
	"github.com/gorilla/securecookie"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
	"github.com/muesli/cache2go"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/srand"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host/abstracts"
	"github.com/syncfuture/host/model"
	"github.com/syncfuture/host/shttp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	_cookieTokenProtectorKey = "token"
	_userJsonSessionkey      = "UserJson"
	_userIDSessionKey        = "userIDSessionKey"
)

type (
	IrisOAuthClient struct {
		IrisBaseServer
		abstracts.OAuthClientHost
		AccessDeniedPath   string
		StaticFilesDir     string
		ViewsExtension     string
		LayoutTemplate     string
		SessionName        string
		TokenCookieName    string
		userJsonSessionkey string
		userIDSessionKey   string
		SessionManager     *sessions.Sessions
		UserLocks          *cache2go.CacheTable
	}
)

func NewIrisOAuthClient(options *OAuthClientOptions) (r *IrisOAuthClient) {
	// create pointer
	r = new(IrisOAuthClient)
	r.ConfigIrisBaseServer(&options.IrisBaseServerOptions)
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

	cookieProtoector := securecookie.New([]byte(options.HashKey), []byte(options.BlockKey))
	if options.ContextTokenStore == nil {
		options.ContextTokenStore = shttp.NewCookieTokenStore(options.TokenCookieName, cookieProtoector)
	}
	if options.OAuthClientHandler == nil {
		options.OAuthClientHandler = abstracts.NewOAuthClientHandler(options.OAuth, options.ContextTokenStore, _userJsonSessionkey, _userIDSessionKey, options.TokenCookieName)
	}

	r.OAuthOptions = options.OAuth
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
	r.userIDSessionKey = _userIDSessionKey
	r.userJsonSessionkey = _userIDSessionKey
	r.CookieProtoector = securecookie.New([]byte(options.HashKey), []byte(options.BlockKey))
	r.SessionManager = sessions.New(sessions.Config{
		Cookie:                      r.SessionName,
		Expires:                     time.Duration(-1),
		Encoding:                    r.CookieProtoector,
		AllowReclaim:                true,
		DisableSubdomainPersistence: true,
	})
	// r.SignInHandler = options.SignInHandler
	// r.SignInCallbackHandler = options.SignInCallbackHandler
	// r.SignOutHandler = options.SignOutHandler
	// r.SignOutCallbackHandler = options.SignOutCallbackHandler
	r.UserLocks = cache2go.Cache("UserLocks")

	// 添加内置终结点
	r.IrisApp.Get(r.SignInPath, AdaptHandler(r.SessionManager, r.OAuthClientHandler.SignInHandler))
	r.IrisApp.Get(r.SignInCallbackPath, AdaptHandler(r.SessionManager, r.OAuthClientHandler.SignInCallbackHandler))
	r.IrisApp.Get(r.SignOutPath, AdaptHandler(r.SessionManager, r.OAuthClientHandler.SignOutHandler))
	r.IrisApp.Get(r.SignOutCallbackPath, AdaptHandler(r.SessionManager, r.OAuthClientHandler.SignOutCallbackHandler))

	// 注册视图引擎
	if r.ViewEngine == nil {
		r.ViewEngine = iris.HTML(r.ViewsDir, r.ViewsExtension).Layout(r.LayoutTemplate).Reload(r.Debug)
	}
	r.IrisApp.RegisterView(r.ViewEngine)

	// 注册静态文件
	r.IrisApp.HandleDir("/", r.StaticFilesDir)

	return
}

func (x *IrisOAuthClient) Run(actionGroups ...*[]*Action) {
	// 构造页面路由字典
	actionMap := make(map[string]*Action)
	for _, actionGroup := range actionGroups {
		for _, action := range *actionGroup {
			actionMap[action.Route] = action
		}
	}
	x.ActionMap = &actionMap

	x.RegisterActions()

	if x.ListenAddr == "" {
		log.Fatal("cannot find 'ListenAddr' in config")
	}
	x.IrisApp.Run(iris.Addr(x.ListenAddr))
}

func (x *IrisOAuthClient) MvcAuthorize(ctx iris.Context) {
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
	if x.OAuthOptions.PkceRequired {
		codeVerifier := oauth2core.Random64String()
		codeChanllenge := oauth2core.ToSHA256Base64URL(codeVerifier)
		session.Set(oauth2core.Form_CodeVerifier, codeVerifier)
		session.Set(oauth2core.Form_CodeChallengeMethod, oauth2core.Pkce_S256)
		codeChanllengeParam := oauth2.SetAuthURLParam(oauth2core.Form_CodeChallenge, codeChanllenge)
		codeChanllengeMethodParam := oauth2.SetAuthURLParam(oauth2core.Form_CodeChallengeMethod, oauth2core.Pkce_S256)
		ctx.Redirect(x.OAuthOptions.AuthCodeURL(state, codeChanllengeParam, codeChanllengeMethodParam), http.StatusFound)
	} else {
		ctx.Redirect(x.OAuthOptions.AuthCodeURL(state), http.StatusFound)
	}
}

func (x *IrisOAuthClient) Client() (*http.Client, error) {
	return x.OAuthOptions.ClientCredential.Client(context.Background()), nil
}

func (x *IrisOAuthClient) GetUserLock(userID string) *sync.RWMutex {
	if !x.UserLocks.Exists(userID) {
		x.UserLocks.Add(userID, time.Second*30, new(sync.RWMutex))
	}

	userLockCache, _ := x.UserLocks.Value(userID)
	return userLockCache.Data().(*sync.RWMutex)
}

func (x *IrisOAuthClient) UserClient(ctx iris.Context) (*http.Client, error) {
	goctx := context.Background()
	userID := x.GetUserID(ctx)
	if userID == "" {
		return http.DefaultClient, nil
	}

	// 获取用户锁
	userLock := x.GetUserLock(userID)

	// read lock
	userLock.RLock()
	t, err := x.ContextTokenStore.GetToken(NewIrisContext(ctx, x.SessionManager))
	userLock.RUnlock()

	if err != nil {
		return http.DefaultClient, err
	}

	tokenSource := x.OAuthOptions.TokenSource(goctx, t)
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
		x.ContextTokenStore.SaveToken(NewIrisContext(ctx, x.SessionManager), newToken)
		// unlock
		defer userLock.Unlock()
	}

	return oauth2.NewClient(goctx, tokenSource), nil
}

func (x *IrisOAuthClient) GetUser(ctx iris.Context) (r *model.User) {
	session := x.SessionManager.Start(ctx)
	userJson := session.GetString(x.userJsonSessionkey)
	if userJson != "" {
		// 已登录
		err := json.Unmarshal([]byte(userJson), &r)
		u.LogError(err)
	}
	return
}

func (x *IrisOAuthClient) GetUserID(ctx iris.Context) string {
	session := x.SessionManager.Start(ctx)
	return session.GetString(x.userIDSessionKey)
}

func (x *IrisOAuthClient) SignOut(ctx iris.Context) {
	x.SessionManager.Destroy(ctx)
	ctx.RemoveCookie(x.TokenCookieName)
}
