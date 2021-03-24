package sfasthttp

import (
	"github.com/syncfuture/go/sconfig"
	"github.com/syncfuture/host/token"
)

type TokenOption func(*FHOAuthTokenHost)

type FHOAuthTokenHost struct {
	token.OAuthTokenHost
	FHWebHost
}

func NewFHOAuthTokenHost(cp sconfig.IConfigProvider, options ...TokenOption) token.IOAuthTokenHost {
	r := new(FHOAuthTokenHost)
	// r.OAuthTokenHost = new(resource.OAuthTokenHost)
	// r.OAuthTokenHost.BaseHost = new(host.BaseHost)
	// r.FHWebHost = new(FHWebHost)
	cp.GetStruct("@this", &r)
	r.ConfigProvider = cp

	for _, o := range options {
		o(r)
	}

	r.BuildFHOAuthTokenHost()

	return r
}

func (x *FHOAuthTokenHost) BuildFHOAuthTokenHost(options ...TokenOption) {
	x.BuildOAuthTokenHost()
	x.buildFHWebHost()
}
