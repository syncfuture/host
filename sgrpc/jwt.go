package sgrpc

import (
	context "context"
	"strings"

	"google.golang.org/grpc/metadata"

	"github.com/pascaldekloe/jwt"
	"google.golang.org/grpc"
)

func AttachJWTToken(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	token := extractToken(ctx)
	if token == "" {
		return handler(ctx, req)
	}

	claims, err := jwt.ParseWithoutCheck([]byte(token)) // todo
	if err == nil {
		return handler(context.WithValue(ctx, "user", claims.Set), req)
	}

	return handler(ctx, req)
}

func extractToken(ctx context.Context) string {
	md, _ := metadata.FromIncomingContext(ctx)
	if authStrs, ok := md[_authHeader]; ok {
		if len(authStrs) == 1 {
			array := strings.Split(authStrs[0], " ")
			if len(array) == 2 {
				return array[1]
			}
		}
	}
	return ""
}
