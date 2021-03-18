package sfasthttp

import (
	"github.com/syncfuture/go/sconfig"
	"github.com/syncfuture/host/abstracts"
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
	x.BuildFHWebHost(x.BaseWebHost)

	////////// oauth client endpoints
	x.Router.GET(x.SignInPath, ToNativeHandler(x.SessionManager, x.OAuthClientHandler.SignInHandler))
	x.Router.GET(x.SignInCallbackPath, ToNativeHandler(x.SessionManager, x.OAuthClientHandler.SignInCallbackHandler))
	x.Router.GET(x.SignOutPath, ToNativeHandler(x.SessionManager, x.OAuthClientHandler.SignOutHandler))
	x.Router.GET(x.SignOutCallbackPath, ToNativeHandler(x.SessionManager, x.OAuthClientHandler.SignOutCallbackHandler))
}
