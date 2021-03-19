package service

import "github.com/syncfuture/host"

type IServiceHost interface {
	host.IHost
	host.IBaseHost
}

type ServiceHost struct {
	host.BaseHost
	ListenAddr string
}

func (x *ServiceHost) BuildServiceHost() {
	x.BaseHost.BuildBaseHost()
}
