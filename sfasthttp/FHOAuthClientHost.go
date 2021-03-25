package sfasthttp

import (
	"github.com/syncfuture/go/sconfig"
	"github.com/syncfuture/host/client"
)

type ClientHostOption func(*FHOAuthClientHost)

type FHOAuthClientHost struct {
	client.OAuthClientHost
	FHWebHost
}

func NewFHOAuthClientHost(cp sconfig.IConfigProvider, options ...ClientHostOption) client.IOAuthClientHost {
	x := new(FHOAuthClientHost)
	cp.GetStruct("@this", &x)
	x.ConfigProvider = cp

	for _, o := range options {
		o(x)
	}

	x.BuildFHOAuthClientHost()

	return x
}

func (x *FHOAuthClientHost) BuildFHOAuthClientHost() {
	x.BuildOAuthClientHost()
	x.FHWebHost.CookieEncryptor = x.SecureCookieHost.GetCookieEncryptor()
	x.FHWebHost.buildFHWebHost()

	////////// oauth client endpoints
	x.Router.GET(x.SignInPath, x.FHWebHost.BuildNativeHandler(x.SignInPath, x.OAuthClientHandler.SignInHandler))
	x.Router.GET(x.SignInCallbackPath, x.FHWebHost.BuildNativeHandler(x.SignInPath, x.OAuthClientHandler.SignInCallbackHandler))
	x.Router.GET(x.SignOutPath, x.FHWebHost.BuildNativeHandler(x.SignInPath, x.OAuthClientHandler.SignOutHandler))
	x.Router.GET(x.SignOutCallbackPath, x.FHWebHost.BuildNativeHandler(x.SignInPath, x.OAuthClientHandler.SignOutCallbackHandler))
}
