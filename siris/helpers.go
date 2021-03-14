package siris

import (
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host/shttp"
)

func AdaptHandler(handler shttp.RequestHandler, sess *sessions.Sessions) iris.Handler {
	return iris.Handler(func(ctx iris.Context) {
		newCtx := NewIrisContext(ctx, sess)
		handler(newCtx)
	})
}

func HandleError(ctx iris.Context, err error) bool {
	if u.LogError(err) {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.WriteString(err.Error())
		return true
	}
	return false
}

func NewAction(route, area, controller string, handler iris.Handler) *Action {
	action := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
	action = action[strings.LastIndex(action, ".")+1:]

	return &Action{
		Route:      route,
		Area:       area,
		Controller: controller,
		Action:     action,
		Handler:    handler,
	}
}
