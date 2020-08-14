package host

import (
	"net"

	"github.com/syncfuture/go/sgrpc"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	panichandler "github.com/kazegusuri/grpc-panic-handler"
	"github.com/syncfuture/go/config"
	log "github.com/syncfuture/go/slog"
	"google.golang.org/grpc"
)

type (
	ServiceServerOptions struct {
		BaseServerOptions
		MaxRecvMsgSize int
	}

	ServiceServer struct {
		BaseServer
		GRPCServer *grpc.Server
	}
)

func NewServiceServerOptions(args ...string) *ServiceServerOptions {
	cp := config.NewJsonConfigProvider(args...)
	var options *ServiceServerOptions
	cp.GetStruct("ServiceServer", &options)
	if options == nil {
		log.Fatal("missing 'ServiceServer' section in configuration")
	}
	options.ConfigProvider = cp
	return options
}

func NewServiceServer(options *ServiceServerOptions) (r *ServiceServer) {
	if options.MaxRecvMsgSize == 0 {
		options.MaxRecvMsgSize = 10 * 1024 * 1024
	}

	r = new(ServiceServer)
	r.Name = options.Name
	r.configBaseServer(&options.BaseServerOptions)

	// GRPC Server
	unaryHandler := grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(panichandler.UnaryPanicHandler, sgrpc.AttachJWTToken))
	streamHandler := grpc.StreamInterceptor(panichandler.StreamPanicHandler)
	panichandler.InstallPanicHandler(func(r interface{}) {
		log.Error(r)
	})
	r.GRPCServer = grpc.NewServer(grpc.MaxRecvMsgSize(options.MaxRecvMsgSize), unaryHandler, streamHandler)

	return r
}

func (x *ServiceServer) Run() {
	lis, err := net.Listen("tcp", x.ListenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Infof("Listening at %v\n", x.ListenAddr)
	if err := x.GRPCServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
