package client

import (
	"github.com/syncfuture/host"
	"golang.org/x/oauth2"
)

type IOAuthClientHost interface {
	host.IBaseHost
	host.IWebHost
	AuthHandler(ctx host.IHttpContext)
	GetClientToken(ctx host.IHttpContext) (*oauth2.Token, error)
	GetUserToken(ctx host.IHttpContext) (*oauth2.TokenSource, error)
}
