package sgrpc

import (
	"context"
	"fmt"

	oauth2go "github.com/Lukiya/oauth2go/core"
	"github.com/pascaldekloe/jwt"
	"github.com/syncfuture/go/sconv"
	"github.com/syncfuture/go/serr"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host"
	_ "github.com/syncfuture/host/sconsul"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	Header_Token = "token"
	Ctx_Claims   = "claims"
)

// func CreateServer() *grpc.Server {
// 	uIntOpt := grpc.UnaryInterceptor(panichandler.UnaryPanicHandler)
// 	sIntOpt := grpc.StreamInterceptor(panichandler.StreamPanicHandler)
// 	panichandler.InstallPanicHandler(func(r interface{}) {
// 		slog.Error(r)
// 	})
// 	return grpc.NewServer(uIntOpt, sIntOpt)
// }

// DialWithHttpContextToken 拨号，发送令牌
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

// receiveTokenMiddleware 接收令牌中间件
func receiveTokenMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	claims, err := receiveTokenMiddleware_ExtractClaims(ctx) // 从收到的令牌中提取出Claims
	if err != nil || claims == nil {
		u.LogError(err)
		return handler(ctx, req)
	}

	// 提取成功，附加给context
	return handler(context.WithValue(ctx, Ctx_Claims, claims), req) // RL00003
}

func receiveTokenMiddleware_ExtractClaims(ctx context.Context) (*map[string]interface{}, error) {
	if metas, ok := metadata.FromIncomingContext(ctx); ok {
		if tokenArray, ok := metas[Header_Token]; ok {
			if len(tokenArray) == 1 {
				claims, err := jwt.ParseWithoutCheck(u.StrToBytes(tokenArray[0]))
				if err != nil {
					return nil, serr.WithStack(err)
				}

				claims.Set[oauth2go.Claim_Subject] = claims.Subject

				return &claims.Set, nil
			}
		}
	}
	return nil, nil
}

func getClaims(ctx context.Context) *map[string]interface{} {
	j, ok := ctx.Value(Ctx_Claims).(*map[string]interface{}) // RL00003

	if ok {
		return j
	}

	return nil
}

func getClaimValue(ctx context.Context, claimName string) interface{} {
	claims := getClaims(ctx)
	if claims != nil {
		if v, ok := (*claims)[claimName]; ok {
			return v
		}
	}
	return nil
}

func GetClaimString(ctx context.Context, claimName string) string {
	v := getClaimValue(ctx, claimName)
	return sconv.ToString(v)
}

func GetClaimInt64(ctx context.Context, claimName string) int64 {
	v := getClaimValue(ctx, claimName)
	return sconv.ToInt64(v)
}

func DialConsul(consulAddr, serviceName string, args map[string]string) (*grpc.ClientConn, error) {
	url := fmt.Sprintf("%s://%s/%s", "consul", consulAddr, serviceName)
	if len(args) > 0 {
		url = url + "?"
		for k, v := range args {
			url = url + k + "=" + v
		}
	}

	r, err := grpc.Dial(
		url,
		//不能block => blockkingPicker打开，在调用轮询时picker_wrapper => picker时若block则不进行robin操作直接返回失败
		//grpc.WithBlock(),
		grpc.WithInsecure(),
		//指定初始化round_robin => balancer (后续可以自行定制balancer和 register、resolver 同样的方式)
		// grpc.WithBalancerName(roundrobin.Name),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
		//grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
	)
	if err != nil {
		return nil, serr.WithStack(err)
	}
	return r, nil
}
