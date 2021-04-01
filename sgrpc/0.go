package sgrpc

import (
	"context"
	"encoding/json"

	"github.com/syncfuture/go/serr"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const (
	CONSTS_USER   = "user"
	_headerClaims = "claims"
)

// func CreateServer() *grpc.Server {
// 	uIntOpt := grpc.UnaryInterceptor(panichandler.UnaryPanicHandler)
// 	sIntOpt := grpc.StreamInterceptor(panichandler.StreamPanicHandler)
// 	panichandler.InstallPanicHandler(func(r interface{}) {
// 		log.Error(r)
// 	})
// 	return grpc.NewServer(uIntOpt, sIntOpt)
// }

func DialWithHttpContextToken(addr string, ctx host.IHttpContext) (r *grpc.ClientConn, err error) {
	j := ctx.GetItem(host.Ctx_Token) // RL00002
	if j != nil {
		token, ok := j.(string)
		if ok {
			r, err = grpc.Dial(
				addr,
				grpc.WithInsecure(),
				grpc.WithPerRPCCredentials(newTokenCredential(token, false)),
			)
		}
	}

	if r == nil {
		r, err = grpc.Dial(
			addr,
			grpc.WithInsecure(),
		)
	}

	return r, serr.WithStack(err)
}

func attachUserClaims(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	claimsJson := extractClaims(ctx)
	if claimsJson == "" {
		return handler(ctx, req)
	}

	// claims, err := jwt.ParseWithoutCheck(u.StrToBytes(claimsJson)) // todo
	var claims map[string]interface{}
	err = json.Unmarshal(u.StrToBytes(claimsJson), &claims)
	if err == nil {
		// return handler(context.WithValue(ctx, "user", claims.Set), req)
		return handler(context.WithValue(ctx, CONSTS_USER, claims), req)
	} else {
		err = serr.WithStack(err)
	}

	return handler(ctx, req)
}

func extractClaims(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if authStrs, ok := md[_headerClaims]; ok {
			if len(authStrs) == 1 {
				return authStrs[0]
			}
		}
	}
	return ""
}

func newTokenCredential(token string, requireTLS bool) credentials.PerRPCCredentials {
	return &tokenCredential{
		ClaimsJson: token,
		RequireTLS: requireTLS,
	}
}
