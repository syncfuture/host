package host

type (
	GrpcServerOptions struct {
		BaseHostOptions
		MaxRecvMsgSize int
	}
)
