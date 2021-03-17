package sfasthttp

import (
	"github.com/syncfuture/go/sconfig"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/host/abstracts"
	"github.com/valyala/fasthttp"
)

type ResourceOption func(*FHOAuthResourceHost)

type FHOAuthResourceHost struct {
	*abstracts.OAuthResourceHost
	*FHWebHost
}

func NewFHOAuthResourceHost(cp sconfig.IConfigProvider, options ...ResourceOption) *FHOAuthResourceHost {
	r := new(FHOAuthResourceHost)
	r.OAuthResourceHost = new(abstracts.OAuthResourceHost)
	r.OAuthResourceHost.BaseHost = new(abstracts.BaseHost)
	r.FHWebHost = new(FHWebHost)
	cp.GetStruct("@this", &r)
	r.ConfigProvider = cp

	for _, o := range options {
		o(r)
	}

	r.BuildFHOAuthResourceHost()

	return r
}

func (x *FHOAuthResourceHost) BuildFHOAuthResourceHost(options ...ResourceOption) {
	x.BuildOAuthResourceHost()
	x.BuildFHWebHost()
}

func (x *FHOAuthResourceHost) Serve() error {
	log.Infof("Listening on %s", x.ListenAddr)
	return fasthttp.ListenAndServe(x.ListenAddr, x.Router.Handler)
}
