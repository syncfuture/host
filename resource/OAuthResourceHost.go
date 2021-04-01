package resource

import (
	"crypto/rsa"
	"net/http"
	"strings"
	"time"

	oauth2core "github.com/Lukiya/oauth2go/core"
	"github.com/Lukiya/oauth2go/model"
	"github.com/pascaldekloe/jwt"
	"github.com/syncfuture/go/sconv"
	"github.com/syncfuture/go/shttp"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/srsautil"
	"github.com/syncfuture/go/sslice"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host"
)

type IOAuthResourceHost interface {
	host.IBaseHost
	host.IWebHost
	AuthHandler(ctx host.IHttpContext)
}

type OAuthResourceHost struct {
	host.BaseHost
	OAuthOptions     *model.Resource `json:"OAuth,omitempty"`
	PublicKey        *rsa.PublicKey
	PublicKeyPath    string
	SigningAlgorithm string
	// TokenValidator   func(*jwt.Claims) string
}

func (x *OAuthResourceHost) BuildOAuthResourceHost() {
	x.BaseHost.BuildBaseHost()

	if x.OAuthOptions == nil {
		log.Fatal("OAuth secion in configuration is missing")
	}

	if x.PublicKeyPath == "" {
		log.Fatal("public key path cannot be empty")
	}
	if x.OAuthOptions == nil {
		log.Fatal("oauth options cannot be nil")
	}
	if x.OAuthOptions.ValidIssuers == nil || len(x.OAuthOptions.ValidIssuers) == 0 {
		log.Fatal("Issuers cannot be empty")
	}
	if x.OAuthOptions.ValidAudiences == nil || len(x.OAuthOptions.ValidAudiences) == 0 {
		log.Fatal("Audiences cannot be empty")
	}
	if x.SigningAlgorithm == "" {
		x.SigningAlgorithm = jwt.PS256
	}

	if x.URLProvider != nil {
		for i := range x.OAuthOptions.ValidIssuers {
			x.OAuthOptions.ValidIssuers[i] = x.URLProvider.RenderURL(x.OAuthOptions.ValidIssuers[i])
		}
		for i := range x.OAuthOptions.ValidAudiences {
			x.OAuthOptions.ValidAudiences[i] = x.URLProvider.RenderURL(x.OAuthOptions.ValidAudiences[i])
		}
	}

	// read public certificate
	cert, err := srsautil.ReadCertFromFile(x.PublicKeyPath)
	u.LogFaltal(err)
	x.PublicKey = cert.PublicKey.(*rsa.PublicKey)
}

func (x *OAuthResourceHost) AuthHandler(ctx host.IHttpContext) {
	authHeader := ctx.GetHeader(shttp.HEADER_AUTH)
	if authHeader == "" {
		ctx.SetStatusCode(http.StatusUnauthorized)
		ctx.WriteString("Authorization header is missing")
		return
	}

	// verify authorization header
	array := strings.Split(authHeader, " ")
	if len(array) != 2 || array[0] != host.AuthType_Bearer {
		ctx.SetStatusCode(http.StatusBadRequest)
		log.Warnf("'%s'invalid authorization header format. '%s'", ctx.GetRemoteIP(), authHeader)
		return
	}
	token := array[1]

	// verify signature
	jwtClaims, err := jwt.RSACheck(u.StrToBytes(token), x.PublicKey)
	if err != nil {
		ctx.SetStatusCode(http.StatusUnauthorized)
		log.Warn("'"+ctx.GetRemoteIP()+"'", err)
		return
	}

	// validate time limits
	isNotExpired := jwtClaims.Valid(time.Now().UTC())
	if !isNotExpired {
		ctx.SetStatusCode(http.StatusUnauthorized)
		msgCode := "current time not in token's valid period"
		ctx.WriteString(msgCode)
		log.Warn("'"+ctx.GetRemoteIP()+"'", msgCode)
		return
	}

	// validate aud
	isValidAudience := x.OAuthOptions.ValidAudiences != nil && sslice.HasAnyStr(x.OAuthOptions.ValidAudiences, jwtClaims.Audiences)
	if !isValidAudience {
		ctx.SetStatusCode(http.StatusUnauthorized)
		msgCode := "invalid audience"
		ctx.WriteString(msgCode)
		log.Warn("'"+ctx.GetRemoteIP()+"'", msgCode)
		return
	}

	// validate iss
	isValidIssuer := x.OAuthOptions.ValidIssuers != nil && sslice.HasStr(x.OAuthOptions.ValidIssuers, jwtClaims.Issuer)
	if !isValidIssuer {
		ctx.SetStatusCode(http.StatusUnauthorized)
		msgCode := "invalid issuer"
		ctx.WriteString(msgCode)
		log.Warn("'"+ctx.GetRemoteIP()+"'", msgCode)
		return
	}

	// if x.TokenValidator != nil {
	// 	if msgCode := x.TokenValidator(token); msgCode != "" {
	// 		ctx.SetStatusCode(http.StatusUnauthorized)
	// 		ctx.WriteString(msgCode)
	// 		log.Warn("'"+ctx.GetRemoteIP()+"'", msgCode)
	// 		return
	// 	}
	// }

	var msgCode string
	if jwtClaims != nil {
		routeKey := ctx.GetItemString(host.Ctx_RouteKey)
		area, controller, action := host.GetRoutesByKey(routeKey)

		roles := sconv.ToInt64(jwtClaims.Set[oauth2core.Claim_Role])
		level := sconv.ToInt64(jwtClaims.Set[oauth2core.Claim_Level])
		if x.PermissionAuditor.CheckRouteWithLevel(area, controller, action, roles, int32(level)) {
			// Has permission, allow
			ctx.SetItem(host.Ctx_Claims, &jwtClaims.Set) // RL00001
			ctx.SetItem(host.Ctx_Token, token)           // RL00002
			ctx.Next()
			return
		} else {
			msgCode = "permission denied"
		}
	}

	// Not allow
	ctx.SetStatusCode(http.StatusUnauthorized)
	ctx.WriteString(msgCode)
}
