package client

import "github.com/syncfuture/host"

type IOAuthClientHost interface {
	host.IBaseHost
	host.IWebHost
}
