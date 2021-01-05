package host

import (
	"crypto/rsa"
	"net/http"
	"strconv"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/syncfuture/go/security"

	"github.com/syncfuture/go/u"

	"github.com/pascaldekloe/jwt"
	log "github.com/syncfuture/go/slog"
)

const (
	_authHeader = "Authorization"
	_jwtValue   = "jwt"
)

/// JWTMiddleware jwt extraction & validation middle ware
type JWTMiddleware struct {
	IssuerSigningKey *rsa.PublicKey
}

func (x *JWTMiddleware) Serve(ctx iris.Context) {
	authHeader := ctx.Request().Header[_authHeader]
	if authHeader == nil || len(authHeader) != 1 {
		ctx.Next()
		return
	}

	array := strings.Split(authHeader[0], " ")
	if len(array) != 2 {
		ctx.Next()
		return
	}

	claims, err := jwt.RSACheck([]byte(array[1]), x.IssuerSigningKey)
	if err != nil {
		log.Warn(err)
		ctx.Next()
		return
	}

	ctx.Values().Set(_jwtValue, claims)
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
	jwtValue := ctx.Values().Get(_jwtValue)
	if jwtValue != nil {
		token := jwtValue.(*jwt.Claims)
		claims := token.Set

		if roleStr, ok := claims["role"].(string); ok && roleStr != "" {
			// Has role filed
			roles, err := strconv.ParseInt(roleStr, 10, 64)
			if !u.LogError(err) {
				// Role can parse to int64
				route := ctx.GetCurrentRoute().Name()
				if action, ok := (*x.ActionMap)[route]; ok {
					// foud action
					var level int
					if levelStr, ok := claims["level"].(string); ok && levelStr != "" {
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
