package sgrpc

import (
	"context"

	"google.golang.org/grpc/credentials"
)

type tokenCredential struct {
	Token      string
	RequireTLS bool
}

func newTokenCredential(token string, requireTLS bool) credentials.PerRPCCredentials {
	return &tokenCredential{
		Token:      token,
		RequireTLS: requireTLS,
	}
}

// GetRequestMetadata 获取请求Meta
func (x *tokenCredential) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	if x.Token != "" {
		return map[string]string{
			// _authHeader: _tokenType + x.Token,
			Header_Token: x.Token,
		}, nil
	}

	return map[string]string{}, nil
}

// RequireTransportSecurity 是否开启TLS
func (x *tokenCredential) RequireTransportSecurity() bool {
	return x.RequireTLS
}
