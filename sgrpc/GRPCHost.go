package sgrpc

import (
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	panichandler "github.com/kazegusuri/grpc-panic-handler"
	"github.com/syncfuture/go/sconfig"
	"github.com/syncfuture/go/serr"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/host/service"
	"google.golang.org/grpc"
)

type GRPCOption func(*GRPCServiceHost)

type IGRPCServiceHost interface {
	service.IServiceHost
	GetGRPCServer() *grpc.Server
}

type GRPCServiceHost struct {
	service.ServiceHost
	GRPCServer     *grpc.Server
	MaxRecvMsgSize int
}

func NewGRPCServiceHost(cp sconfig.IConfigProvider, options ...GRPCOption) IGRPCServiceHost {
	x := new(GRPCServiceHost)
	cp.GetStruct("@this", &x)
	x.ConfigProvider = cp

	for _, o := range options {
		o(x)
	}

	x.BuildGRPCServiceHost()

	return x
}

func (x *GRPCServiceHost) BuildGRPCServiceHost() {
	x.ServiceHost.BuildServiceHost()

	if x.MaxRecvMsgSize == 0 {
		x.MaxRecvMsgSize = 10 * 1024 * 1024
	}

	// GRPC Server
	unaryHandler := grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(panichandler.UnaryPanicHandler, AttachUserClaims))
	streamHandler := grpc.StreamInterceptor(panichandler.StreamPanicHandler)
	panichandler.InstallPanicHandler(func(r interface{}) {
		log.Error(r)
	})

	x.GRPCServer = grpc.NewServer(grpc.MaxRecvMsgSize(x.MaxRecvMsgSize), unaryHandler, streamHandler)
}

func (x *GRPCServiceHost) GetGRPCServer() *grpc.Server {
	return x.GRPCServer
}

func (x *GRPCServiceHost) Run() error {
	listen, err := net.Listen("tcp", x.ListenAddr)
	if err != nil {
		return serr.WithStack(err)
	}

	log.Infof("Listening at %v\n", x.ListenAddr)
	return serr.WithStack(x.GRPCServer.Serve(listen))
}
