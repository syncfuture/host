package sgrpc

import (
	"context"

	"google.golang.org/grpc/credentials"
)

const (
	// _headerClaims = "authorization"
	_headerClaims = "claims"
	// _tokenType  = "Bearer "
)

type tokenCredential struct {
	ClaimsJson string
	RequireTLS bool
}

// GetRequestMetadata 获取请求Meta
func (x tokenCredential) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	if x.ClaimsJson != "" {
		return map[string]string{
			// _authHeader: _tokenType + x.Token,
			_headerClaims: x.ClaimsJson,
		}, nil
	} else {
		return map[string]string{}, nil
	}
}

// RequireTransportSecurity 是否开启TLS
func (x tokenCredential) RequireTransportSecurity() bool {
	return x.RequireTLS
}

func NewTokenCredential(token string, requireTLS bool) credentials.PerRPCCredentials {
	return &tokenCredential{
		ClaimsJson: token,
		RequireTLS: requireTLS,
	}
}
