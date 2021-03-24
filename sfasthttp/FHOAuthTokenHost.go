package sfasthttp

import (
	"github.com/syncfuture/go/sconfig"
	"github.com/syncfuture/host/token"
)

type TokenHostOption func(*FHOAuthTokenHost)

type FHOAuthTokenHost struct {
	token.OAuthTokenHost
	FHWebHost
}

func NewFHOAuthTokenHost(cp sconfig.IConfigProvider, options ...TokenHostOption) token.IOAuthTokenHost {
	r := new(FHOAuthTokenHost)
	cp.GetStruct("@this", &r)
	r.ConfigProvider = cp

	for _, o := range options {
		o(r)
	}

	r.BuildFHOAuthTokenHost()

	return r
}

func (x *FHOAuthTokenHost) BuildFHOAuthTokenHost() {
	x.BuildOAuthTokenHost()
	x.buildFHWebHost()

	x.Router.POST(x.TokenEndpoint, x.TokenHost.TokenRequestHandler)
	x.Router.GET(x.AuthorizeEndpoint, x.TokenHost.AuthorizeRequestHandler)
	x.Router.GET(x.EndSessionEndpoint, x.TokenHost.EndSessionRequestHandler)
	x.Router.POST(x.EndSessionEndpoint, x.TokenHost.ClearTokenRequestHandler)
}
