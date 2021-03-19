package client

import (
	"sync"

	"github.com/syncfuture/host"
	"golang.org/x/oauth2"
)

type IOAuthClientHost interface {
	host.IBaseHost
	host.IWebHost
	AuthHandler(ctx host.IHttpContext)
	GetClientToken(ctx host.IHttpContext) (*oauth2.Token, error)
	GetUserToken(ctx host.IHttpContext) (*oauth2.TokenSource, error)
	GetUserLock(userID string) *sync.RWMutex
	GetUserJsonSessionKey() string
	GetUserIDSessionKey() string
}
