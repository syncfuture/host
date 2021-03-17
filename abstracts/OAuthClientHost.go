package abstracts

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/muesli/cache2go"
	log "github.com/syncfuture/go/slog"
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
		authMiddleware      IAuthMiddleware
		ContextTokenStore   shttp.IContextTokenStore
		CookieProtoector    *securecookie.SecureCookie
		UserLocks           *cache2go.CacheTable
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

	////////// user locks
	if x.UserLocks == nil {
		x.UserLocks = cache2go.Cache("UserLocks")
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
		x.OAuthClientHandler = NewOAuthClientHandler(x.OAuthOptions, x.ContextTokenStore, x.UserJsonSessionKey, x.UserIDSessionKey, x.TokenCookieName)
	}

	////////// auth middleware
	if x.authMiddleware == nil {
		x.authMiddleware = newClientAuthMiddleware(x.UserJsonSessionKey, x.AccessDeniedPath, x.OAuthOptions, x.PermissionAuditor)
	}
}

func (x *OAuthClientHost) GetHttpClient() (*http.Client, error) {
	return x.OAuthOptions.ClientCredential.Client(context.Background()), nil
}

func (x *OAuthClientHost) GetUserHttpClient(ctx shttp.IHttpContext) (*http.Client, error) {
	// goctx := context.Background()
	// userID := shttp.GetUserID(ctx, x.UserIDSessionKey)
	// if userID == "" {
	// 	return http.DefaultClient, nil
	// }

	// // 获取用户锁
	// userLock := x.getUserLock(userID)

	// // read lock
	// userLock.RLock()
	// t, err := x.ContextTokenStore.GetToken(ctx)
	// defer func() { userLock.RUnlock() }()
	// if err != nil {
	// 	return http.DefaultClient, err
	// }

	// tokenSource := x.OAuthOptions.TokenSource(goctx, t)
	// newToken, err := tokenSource.Token()
	// if err != nil {
	// 	// refresh token failed, sign user out
	// 	shttp.SignOut(ctx, x.TokenCookieName)
	// 	return http.DefaultClient, err
	// }

	// if newToken.AccessToken != t.AccessToken {
	// 	// token been refreshed, lock
	// 	userLock.Lock()
	// 	// save token to session
	// 	err = x.ContextTokenStore.SaveToken(ctx, newToken)
	// 	// unlock
	// 	defer func() { userLock.Unlock() }()
	// 	if err != nil {
	// 		return http.DefaultClient, err
	// 	}
	// }

	tokenSource, err := x.GetUserToken(ctx)
	if err != nil {
		return nil, err
	}
	return oauth2.NewClient(context.Background(), *tokenSource), nil
}

func (x *OAuthClientHost) GetClientToken(ctx shttp.IHttpContext) (*oauth2.Token, error) {
	return x.OAuthOptions.ClientCredential.Token()
}

func (x *OAuthClientHost) GetUserToken(ctx shttp.IHttpContext) (*oauth2.TokenSource, error) {
	goctx := context.Background()
	userID := shttp.GetUserID(ctx, x.UserIDSessionKey)
	if userID == "" {
		return nil, errors.New("user isn't authenticated")
	}

	// 获取用户锁
	userLock := x.getUserLock(userID)

	// read lock
	userLock.RLock()
	t, err := x.ContextTokenStore.GetToken(ctx)
	defer func() { userLock.RUnlock() }()
	if err != nil {
		return nil, err
	}

	tokenSource := x.OAuthOptions.TokenSource(goctx, t)
	newToken, err := tokenSource.Token()
	if err != nil {
		// refresh token failed, sign user out
		shttp.SignOut(ctx, x.TokenCookieName)
		return nil, err
	}

	if newToken.AccessToken != t.AccessToken {
		// token been refreshed, lock
		userLock.Lock()
		// save token to session
		err = x.ContextTokenStore.SaveToken(ctx, newToken)
		// unlock
		defer func() { userLock.Unlock() }()
		if err != nil {
			return nil, err
		}
	}

	return &tokenSource, nil
}

func (x *OAuthClientHost) Auth(next shttp.RequestHandler, routes ...string) shttp.RequestHandler {
	return x.authMiddleware.Serve(next, routes...)
}

func (x *OAuthClientHost) getUserLock(userID string) *sync.RWMutex {
	if !x.UserLocks.Exists(userID) {
		x.UserLocks.Add(userID, time.Second*30, new(sync.RWMutex))
	}

	userLockCache, err := x.UserLocks.Value(userID)
	u.LogError(err)
	return userLockCache.Data().(*sync.RWMutex)
}
