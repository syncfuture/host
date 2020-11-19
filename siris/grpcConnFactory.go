package siris

import (
	"github.com/iris-contrib/middleware/jwt"
	"github.com/kataras/iris/v12"
	"github.com/syncfuture/host/sgrpc"
	"google.golang.org/grpc"

	log "github.com/syncfuture/go/slog"
)

func CreateGRPCConnPool(ctx iris.Context, addr string) (r *grpc.ClientConn) {
	var err error
	j := ctx.Values().Get("jwt")
	if j != nil {
		token, ok := j.(*jwt.Token)
		if ok {
			r, err = grpc.Dial(
				addr,
				grpc.WithInsecure(),
				grpc.WithPerRPCCredentials(sgrpc.NewTokenCredential(token.Raw, false)),
			)
		}
	}

	if r == nil {
		r, err = grpc.Dial(
			addr,
			grpc.WithInsecure(),
		)
	}

	if err != nil {
		log.Fatal(err)
	}

	return r
}
