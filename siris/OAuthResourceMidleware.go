package siris

import (
	"crypto/rsa"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	security "github.com/syncfuture/go/ssecurity"
	"github.com/syncfuture/go/sslice"

	"github.com/syncfuture/go/u"

	oauth2core "github.com/Lukiya/oauth2go/core"
	"github.com/pascaldekloe/jwt"
	log "github.com/syncfuture/go/slog"
)

const (
	_authHeader = "Authorization"
	_jwtValue   = "jwt"
	_bearer     = "Bearer"
)

/// JWTMiddleware jwt extraction & validation middle ware
type JWTMiddleware struct {
	IssuerSigningKey *rsa.PublicKey
	ValidAudiences   []string
	ValidIssuers     []string
	TokenValidator   func(*jwt.Claims) string
}

func (x *JWTMiddleware) Serve(ctx iris.Context) {
	authHeader := ctx.Request().Header[_authHeader]
	if authHeader == nil || len(authHeader) != 1 {
		ctx.Next()
		return
	}

	// verify authorization header
	array := strings.Split(authHeader[0], " ")
	if len(array) != 2 || array[0] != _bearer {
		ctx.StatusCode(http.StatusBadRequest)
		log.Warnf("'%s'invalid authorization header format. '%s'", ctx.Request().RemoteAddr, authHeader[0])
		return
	}

	// verify signature
	token, err := jwt.RSACheck([]byte(array[1]), x.IssuerSigningKey)
	if err != nil {
		ctx.StatusCode(http.StatusUnauthorized)
		log.Warn("'"+ctx.Request().RemoteAddr+"'", err)
		return
	}

	// validate time limits
	isNotExpired := token.Valid(time.Now().UTC())
	if !isNotExpired {
		ctx.StatusCode(http.StatusUnauthorized)
		msgCode := "current time not in token's valid period"
		ctx.WriteString(msgCode)
		log.Warn("'"+ctx.Request().RemoteAddr+"'", msgCode)
		return
	}

	// validate aud
	isValidAudience := x.ValidAudiences != nil && sslice.HasAnyStr(x.ValidAudiences, token.Audiences)
	if !isValidAudience {
		ctx.StatusCode(http.StatusUnauthorized)
		msgCode := "invalid audience"
		ctx.WriteString(msgCode)
		log.Warn("'"+ctx.Request().RemoteAddr+"'", msgCode)
		return
	}

	// validate iss
	isValidIssuer := x.ValidIssuers != nil && sslice.HasStr(x.ValidIssuers, token.Issuer)
	if !isValidIssuer {
		ctx.StatusCode(http.StatusUnauthorized)
		msgCode := "invalid issuer"
		ctx.WriteString(msgCode)
		log.Warn("'"+ctx.Request().RemoteAddr+"'", msgCode)
		return
	}

	if x.TokenValidator != nil {
		if msgCode := x.TokenValidator(token); msgCode != "" {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.WriteString(msgCode)
			log.Warn("'"+ctx.Request().RemoteAddr+"'", msgCode)
			return
		}
	}

	ctx.Values().Set(_jwtValue, token)
	ctx.Next()
}

/// ApiAuthMidleware api authorization middle ware
type ApiAuthMidleware struct {
	PermissionAuditor security.IPermissionAuditor
	ActionMap         *map[string]*Action
	ProjectName       string
}

func (x *ApiAuthMidleware) Serve(ctx iris.Context) {
	var msgCode string
	token := ctx.Values().Get(_jwtValue)
	if token != nil {
		claims := token.(*jwt.Claims).Set

		if roleStr, ok := claims[oauth2core.Claim_Role].(string); ok && roleStr != "" {
			// Has role filed
			roles, err := strconv.ParseInt(roleStr, 10, 64)
			if !u.LogError(err) {
				// Role can parse to int64
				route := ctx.GetCurrentRoute().Name()
				if action, ok := (*x.ActionMap)[route]; ok {
					// foud action
					var level int
					if levelStr, ok := claims[oauth2core.Claim_Level].(string); ok && levelStr != "" {
						level, err = strconv.Atoi(levelStr)
						u.LogError(err)
					}
					if x.PermissionAuditor.CheckRouteWithLevel(action.Area, action.Controller, action.Action, roles, int32(level)) {
						// Has permission, allow
						ctx.Next()
						return
					} else {
						msgCode = "permission denied"
					}
				} else {
					msgCode = route + " doesn't exist in action map"
					log.Warn(msgCode)
				}
			} else {
				msgCode = "parse role error"
			}
		} else {
			msgCode = "token doesn't have role field"
			log.Warn(msgCode, " ", claims)
		}
	}

	// Not allow
	ctx.StatusCode(http.StatusUnauthorized)
	ctx.WriteString(msgCode)
}
