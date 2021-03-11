package sgrpc

import (
	context "context"
	"encoding/json"

	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc"
)

const (
	USER = "user"
)

func AttachUserClaims(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	claimsJson := extractClaims(ctx)
	if claimsJson == "" {
		return handler(ctx, req)
	}

	// claims, err := jwt.ParseWithoutCheck([]byte(claimsJson)) // todo
	var claims map[string]interface{}
	err = json.Unmarshal([]byte(claimsJson), &claims)
	if err == nil {
		// return handler(context.WithValue(ctx, "user", claims.Set), req)
		return handler(context.WithValue(ctx, USER, claims), req)
	}

	return handler(ctx, req)
}

func extractClaims(ctx context.Context) string {
	md, _ := metadata.FromIncomingContext(ctx)
	if authStrs, ok := md[_headerClaims]; ok {
		if len(authStrs) == 1 {
			return authStrs[0]
			// array := strings.Split(authStrs[0], " ")
			// if len(array) == 2 {
			// 	return array[1]
			// }
		}
	}
	return ""
}

// func extractToken(ctx context.Context) string {
// 	md, _ := metadata.FromIncomingContext(ctx)
// 	if authStrs, ok := md[_authHeader]; ok {
// 		if len(authStrs) == 1 {
// 			array := strings.Split(authStrs[0], " ")
// 			if len(array) == 2 {
// 				return array[1]
// 			}
// 		}
// 	}
// 	return ""
// }
