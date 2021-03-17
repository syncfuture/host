package sfasthttp

import (
	"github.com/syncfuture/go/sconfig"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/host/abstracts"
	"github.com/valyala/fasthttp"
)

type ClientOption func(*FHOAuthClientHost)

type FHOAuthClientHost struct {
	*abstracts.OAuthClientHost
	*FHWebHost
}

func NewFHOAuthClientHost(cp sconfig.IConfigProvider, options ...ClientOption) *FHOAuthClientHost {
	r := new(FHOAuthClientHost)
	r.OAuthClientHost = new(abstracts.OAuthClientHost)
	r.OAuthClientHost.BaseHost = new(abstracts.BaseHost)
	r.FHWebHost = new(FHWebHost)
	cp.GetStruct("@this", &r)
	r.ConfigProvider = cp

	for _, o := range options {
		o(r)
	}

	r.BuildFHOAuthClientHost()

	return r
}

func (x *FHOAuthClientHost) BuildFHOAuthClientHost(options ...ClientOption) {
	x.BuildOAuthClientHost()
	x.BuildFHWebHost()

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
