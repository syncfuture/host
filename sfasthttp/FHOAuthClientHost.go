package sfasthttp

import (
	"github.com/fasthttp/router"
	"github.com/fasthttp/session/v2"
	"github.com/fasthttp/session/v2/providers/memory"
	"github.com/gorilla/securecookie"
	"github.com/syncfuture/go/sconfig"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host/abstracts"
	"github.com/syncfuture/host/client"
	"github.com/syncfuture/host/shttp"
	"github.com/valyala/fasthttp"
)

const (
	_userJsonSessionkey = "UserJson"
	_userIDSessionKey   = "userIDSessionKey"
	_tokenCookieName    = "go.cookie2"
)

type Option func(*FHOAuthClientHost)

type FHOAuthClientHost struct {
	*abstracts.OAuthClientHost
	*FHWebHost
}

func NewFHOAuthClientHost(cp sconfig.IConfigProvider, options ...Option) *FHOAuthClientHost {
	r := new(FHOAuthClientHost)
	cp.GetStruct("@this", &r)
	r.ConfigProvider = cp
	r.FHWebHost = new(FHWebHost)

	for _, o := range options {
		o(r)
	}

	r.BuildFHOAuthClientHost()

	return r
}

func (x *FHOAuthClientHost) BuildFHOAuthClientHost(options ...Option) {
	x.BuildOAuthClientHost()

	////////// router
	if x.Router == nil {
		x.Router = router.New()
	}

	////////// session provider
	if x.SessionProvider == nil {
		provider, err := memory.New(memory.Config{})
		u.LogFaltal(err)
		x.SessionProvider = provider
	}

	////////// session manager
	if x.SessionManager == nil {
		cfg := session.NewDefaultConfig()
		cfg.EncodeFunc = session.MSGPEncode // 内存型provider性能较好
		cfg.DecodeFunc = session.MSGPDecode // 内存型provider性能较好

		x.SessionManager = session.New(cfg)
		err := x.SessionManager.SetProvider(x.SessionProvider)
		u.LogFaltal(err)
	}

	////////// cookie protoector
	if x.CookieProtoector == nil {
		x.CookieProtoector = securecookie.New(u.StrToBytes(x.HashKey), u.StrToBytes(x.BlockKey))
	}

	////////// context token store
	if x.ContextTokenStore == nil {
		x.ContextTokenStore = shttp.NewCookieTokenStore(_tokenCookieName, x.CookieProtoector)
	}

	////////// oauth client handler
	if x.OAuthClientHandler == nil {
		x.OAuthClientHandler = client.NewDefaultOAuthClientHandler(x.OAuthOptions, x.ContextTokenStore, _userJsonSessionkey, _userIDSessionKey, _tokenCookieName)
	}

	////////// oauth client endpoints
	x.Router.GET(x.SignInPath, AdaptHandler(x.OAuthClientHandler.SignInHandler, x.SessionManager))
	x.Router.GET(x.SignInCallbackPath, AdaptHandler(x.OAuthClientHandler.SignInCallbackHandler, x.SessionManager))
	x.Router.GET(x.SignOutPath, AdaptHandler(x.OAuthClientHandler.SignOutHandler, x.SessionManager))
	x.Router.GET(x.SignOutCallbackPath, AdaptHandler(x.OAuthClientHandler.SignOutCallbackHandler, x.SessionManager))
}

func (x *FHOAuthClientHost) Serve() error {
	log.Infof("Listening on %s", x.ListenAddr)
	return fasthttp.ListenAndServe(x.ListenAddr, x.Router.Handler)
}
