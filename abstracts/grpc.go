package abstracts

type (
	GrpcServerOptions struct {
		BaseServerOptions
		MaxRecvMsgSize int
	}
)
