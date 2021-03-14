package host

import (
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	panichandler "github.com/kazegusuri/grpc-panic-handler"
	config "github.com/syncfuture/go/sconfig"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/host/abstracts"
	"github.com/syncfuture/host/sgrpc"
	"google.golang.org/grpc"
)

type (
	GrpcServer struct {
		abstracts.BaseServer
		GRPCServer *grpc.Server
	}
)

func NewGrpcServerOptions(args ...string) *abstracts.GrpcServerOptions {
	cp := config.NewJsonConfigProvider(args...)
	var options *abstracts.GrpcServerOptions
	cp.GetStruct("GrpcServer", &options)
	if options == nil {
		log.Fatal("missing 'GrpcServer' section in configuration")
	}
	options.ConfigProvider = cp
	return options
}

func NewGrpcServer(options *abstracts.GrpcServerOptions) (r *GrpcServer) {
	if options.MaxRecvMsgSize == 0 {
		options.MaxRecvMsgSize = 10 * 1024 * 1024
	}

	r = new(GrpcServer)
	r.Name = options.Name
	r.URIKey = options.URIKey
	// r.RouteKey = options.RouteKey
	r.PermissionKey = options.PermissionKey
	r.ConfigBaseServer(&options.BaseServerOptions)

	// GRPC Server
	unaryHandler := grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(panichandler.UnaryPanicHandler, sgrpc.AttachUserClaims))
	streamHandler := grpc.StreamInterceptor(panichandler.StreamPanicHandler)
	panichandler.InstallPanicHandler(func(r interface{}) {
		log.Error(r)
	})
	r.GRPCServer = grpc.NewServer(grpc.MaxRecvMsgSize(options.MaxRecvMsgSize), unaryHandler, streamHandler)

	return r
}

func (x *GrpcServer) Run() {
	lis, err := net.Listen("tcp", x.ListenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Infof("Listening at %v\n", x.ListenAddr)
	if err := x.GRPCServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
