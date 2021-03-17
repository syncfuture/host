package abstracts

import (
	"crypto/rsa"
	"net/http"
	"strconv"
	"strings"
	"time"

	oauth2core "github.com/Lukiya/oauth2go/core"
	"github.com/pascaldekloe/jwt"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/ssecurity"
	"github.com/syncfuture/go/sslice"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host/shttp"
)

type JWTAuthMiddleware struct {
	IssuerSigningKey  *rsa.PublicKey
	ValidAudiences    []string
	ValidIssuers      []string
	TokenValidator    func(*jwt.Claims) string
	PermissionAuditor ssecurity.IPermissionAuditor
}

func newJWTAuthMiddleware(
	issuerSigningKey *rsa.PublicKey,
	validAudiences []string,
	validIssuers []string,
	tokenValidator func(*jwt.Claims) string,
	permissionAuditor ssecurity.IPermissionAuditor,
) IAuthMiddleware {
	return &JWTAuthMiddleware{
		IssuerSigningKey:  issuerSigningKey,
		ValidAudiences:    validAudiences,
		ValidIssuers:      validIssuers,
		TokenValidator:    tokenValidator,
		PermissionAuditor: permissionAuditor,
	}
}

func (x *JWTAuthMiddleware) Serve(next shttp.RequestHandler, routes ...string) shttp.RequestHandler {
	var area, controller, action string
	count := len(routes)
	if count == 0 || count > 3 {
		log.Fatal("invalid routes array")
	}

	area = routes[0]
	if count >= 2 {
		controller = routes[1]
	}
	if count == 3 {
		action = routes[2]
	}

	return func(ctx shttp.IHttpContext) {
		authHeader := ctx.GetHeader(shttp.Header_Auth)
		if authHeader == "" {
			next(ctx)
			return
		}

		// verify authorization header
		array := strings.Split(authHeader, " ")
		if len(array) != 2 || array[0] != shttp.AuthType_Bearer {
			ctx.SetStatusCode(http.StatusBadRequest)
			log.Warnf("'%s'invalid authorization header format. '%s'", ctx.GetRemoteIP(), authHeader)
			return
		}

		// verify signature
		token, err := jwt.RSACheck([]byte(array[1]), x.IssuerSigningKey)
		if err != nil {
			ctx.SetStatusCode(http.StatusUnauthorized)
			log.Warn("'"+ctx.GetRemoteIP()+"'", err)
			return
		}

		// validate time limits
		isNotExpired := token.Valid(time.Now().UTC())
		if !isNotExpired {
			ctx.SetStatusCode(http.StatusUnauthorized)
			msgCode := "current time not in token's valid period"
			ctx.WriteString(msgCode)
			log.Warn("'"+ctx.GetRemoteIP()+"'", msgCode)
			return
		}

		// validate aud
		isValidAudience := x.ValidAudiences != nil && sslice.HasAnyStr(x.ValidAudiences, token.Audiences)
		if !isValidAudience {
			ctx.SetStatusCode(http.StatusUnauthorized)
			msgCode := "invalid audience"
			ctx.WriteString(msgCode)
			log.Warn("'"+ctx.GetRemoteIP()+"'", msgCode)
			return
		}

		// validate iss
		isValidIssuer := x.ValidIssuers != nil && sslice.HasStr(x.ValidIssuers, token.Issuer)
		if !isValidIssuer {
			ctx.SetStatusCode(http.StatusUnauthorized)
			msgCode := "invalid issuer"
			ctx.WriteString(msgCode)
			log.Warn("'"+ctx.GetRemoteIP()+"'", msgCode)
			return
		}

		if x.TokenValidator != nil {
			if msgCode := x.TokenValidator(token); msgCode != "" {
				ctx.SetStatusCode(http.StatusUnauthorized)
				ctx.WriteString(msgCode)
				log.Warn("'"+ctx.GetRemoteIP()+"'", msgCode)
				return
			}
		}

		var msgCode string
		if token != nil {
			claims := token.Set

			if roleStr, ok := claims[oauth2core.Claim_Role].(string); ok && roleStr != "" {
				// Has role filed
				roles, err := strconv.ParseInt(roleStr, 10, 64)
				if !u.LogError(err) {
					// Role can parse to int64
					var level int
					if levelStr, ok := claims[oauth2core.Claim_Level].(string); ok && levelStr != "" {
						level, err = strconv.Atoi(levelStr)
						u.LogError(err)
					}
					if x.PermissionAuditor.CheckRouteWithLevel(area, controller, action, roles, int32(level)) {
						// Has permission, allow
						ctx.SetItem(shttp.Item_JWT, token)
						next(ctx)
						return
					} else {
						msgCode = "permission denied"
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
		ctx.SetStatusCode(http.StatusUnauthorized)
		ctx.WriteString(msgCode)
	}
}
