package sfasthttp

import (
	"github.com/syncfuture/go/sconfig"
	"github.com/syncfuture/host/abstracts"
	"github.com/syncfuture/host/resource"
)

type ResourceOption func(*FHOAuthResourceHost)

type FHOAuthResourceHost struct {
	resource.OAuthResourceHost
	FHWebHost
}

func NewFHOAuthResourceHost(cp sconfig.IConfigProvider, options ...ResourceOption) abstracts.IWebHost {
	r := new(FHOAuthResourceHost)
	// r.OAuthResourceHost = new(resource.OAuthResourceHost)
	// r.OAuthResourceHost.BaseHost = new(abstracts.BaseHost)
	// r.FHWebHost = new(FHWebHost)
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
	x.buildFHWebHost()
}
