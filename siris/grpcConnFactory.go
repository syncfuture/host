package siris

import (
	"github.com/kataras/iris/v12"
	"github.com/pascaldekloe/jwt"
	"github.com/syncfuture/host/sgrpc"
	"google.golang.org/grpc"

	log "github.com/syncfuture/go/slog"
)

func CreateGRPCConnPool(ctx iris.Context, addr string) (r *grpc.ClientConn) {
	var err error
	j := ctx.Values().Get("jwt")
	if j != nil {
		token, ok := j.(*jwt.Claims)
		if ok {
			r, err = grpc.Dial(
				addr,
				grpc.WithInsecure(),
				grpc.WithPerRPCCredentials(sgrpc.NewTokenCredential(string(token.Raw), false)),
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
