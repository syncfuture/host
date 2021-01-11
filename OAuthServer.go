package host

import (
	"github.com/Lukiya/oauth2go"
	oauth2core "github.com/Lukiya/oauth2go/core"
	"github.com/Lukiya/oauth2go/security/rsa"
	"github.com/Lukiya/oauth2go/server"
	"github.com/Lukiya/oauth2go/store/redis"
	"github.com/gorilla/securecookie"
	"github.com/syncfuture/go/config"
	"github.com/syncfuture/go/rsautil"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/surl"
	"github.com/syncfuture/go/u"
	"github.com/valyala/fasthttp"
)

type (
	AuthServerOptions struct {
		BaseServerOptions
		oauth2go.AuthServerOptions
		PrivateKeyPath string
		HashKey        string
		BlockKey       string
		ClientStoreKey string
		TokenStoreKey  string
		WebRoot        string
	}

	OAuthServer struct {
		oauth2go.IAuthServer
		server.IWebServer
		BaseServer
		ListenAddr string
	}
)

func NewAuthServerOptions(args ...string) *AuthServerOptions {
	cp := config.NewJsonConfigProvider(args...)
	var options *AuthServerOptions
	cp.GetStruct("OAuthServer", &options)
	if options == nil {
		log.Fatal("missing 'OAuthServer' section in configuration")
	}
	options.ConfigProvider = cp
	return options
}
func NewOAuthServer(options *AuthServerOptions) (r *OAuthServer) {
	options.ConfigProvider.GetStruct("Redis", &options.RedisConfig)
	if options.RedisConfig == nil {
		log.Fatal("missing 'Redis' section in configuration")
	}
	if options.URIKey == "" {
		log.Fatal("missing 'URIKey' in configuration")
	}
	options.URLProvider = surl.NewRedisURLProvider(options.URIKey, options.RedisConfig)

	if options.PrivateKeyPath == "" {
		log.Fatal("missing 'PrivateKeyPath' in configuration")
	}
	var err error
	options.PrivateKey, err = rsautil.ReadPrivateKeyFromFile(options.PrivateKeyPath)
	u.LogFaltal(err)
	secretEncryptor := rsa.NewRSASecretEncryptor(options.PrivateKeyPath)

	if options.HashKey == "" {
		log.Fatal("missing 'HashKey' in configuration")
	}
	if options.BlockKey == "" {
		log.Fatal("missing 'BlockKey' in configuration")
	}
	options.CookieManager = securecookie.New([]byte(options.HashKey), []byte(options.BlockKey))

	if options.ClientStoreKey == "" {
		log.Fatal("missing 'ClientStoreKey' in configuration")
	}
	options.ClientStore = redis.NewRedisClientStore(options.ClientStoreKey, secretEncryptor, options.RedisConfig)

	if options.TokenStoreKey == "" {
		log.Fatal("missing 'TokenStoreKey' in configuration")
	}
	options.TokenStore = redis.NewRedisTokenStore(options.TokenStoreKey, secretEncryptor, options.RedisConfig)

	r = new(OAuthServer)
	r.IAuthServer = oauth2go.NewDefaultAuthServer(&options.AuthServerOptions)
	r.RedisConfig = options.RedisConfig
	r.ConfigProvider = options.ConfigProvider
	r.URLProvider = options.URLProvider
	r.ListenAddr = options.ListenAddr
	r.IWebServer = server.NewWebServer()

	// authorize
	r.Get(options.AuthorizeEndpoint, r.AuthorizeRequestHandler)
	r.Post(options.TokenEndpoint, r.TokenRequestHandler)
	// end session
	r.Get(options.EndSessionEndpoint, r.EndSessionRequestHandler)
	r.Post(options.EndSessionEndpoint, r.ClearTokenRequestHandler)

	// logout
	r.Get(options.LogoutEndpoint, func(ctx *fasthttp.RequestCtx) {
		r.DelCookie(ctx, options.AuthCookieName)
		oauth2core.Redirect(ctx, "/")
	})
	// static files
	r.ServeFiles(fasthttp.FSHandler(options.WebRoot, 0))

	return
}

func (x *OAuthServer) Run() {
	log.Infof("listen on %s", x.ListenAddr)
	fasthttp.ListenAndServe(x.ListenAddr, x.Serve)
}
