package host

import (
	"net/http"
	"strconv"

	"github.com/syncfuture/go/security"

	"github.com/syncfuture/go/u"

	log "github.com/syncfuture/go/slog"

	"github.com/iris-contrib/middleware/jwt"
	iriscontext "github.com/kataras/iris/v12/context"
)

type ApiAuthMidleware struct {
	PermissionAuditor security.IPermissionAuditor
	ActionMap         *map[string]*Action
	ProjectName       string
}

func (x *ApiAuthMidleware) Serve(ctx iriscontext.Context) {
	var msgCode string
	token := ctx.Values().Get("jwt").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)

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

	// Not allow
	ctx.StatusCode(http.StatusUnauthorized)
	ctx.WriteString(msgCode)
}
