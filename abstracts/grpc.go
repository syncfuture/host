package abstracts

type (
	GrpcServerOptions struct {
		BaseHostOptions
		MaxRecvMsgSize int
	}
)
