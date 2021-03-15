package siris

import (
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
	"github.com/pascaldekloe/jwt"
	"github.com/syncfuture/go/sproto/timestamp"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host/shttp"
)

func AdaptHandler(handler shttp.RequestHandler, sess *sessions.Sessions) iris.Handler {
	return iris.Handler(func(ctx iris.Context) {
		var newCtx shttp.IHttpContext
		defer func() {
			PutIrisContext(newCtx)
		}()

		newCtx = NewIrisContext(ctx, sess)
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

func TimeStamp(t *timestamp.Timestamp, args ...interface{}) string {
	utcTime, _ := t.Time()
	beijingTime := utcTime.Add(8 * time.Hour)
	if len(args) > 0 {
		return beijingTime.Format(args[0].(string))
	} else {
		return beijingTime.Format("01/02/2006 03:04:05 PM")
	}
}

func Iter(count int32) []int32 {
	var i int32
	var r []int32
	for i = 0; i < count; i++ {
		r = append(r, i)
	}
	return r
}

func GetClaimInt64(claimName string, ctx iris.Context) int64 {
	str := GetClaimString(claimName, ctx)
	r, _ := strconv.ParseInt(str, 10, 64)
	return r
}

func GetClaimString(claimName string, ctx iris.Context) string {
	j := ctx.Values().Get("jwt")
	if j != nil {
		if token, ok := j.(*jwt.Claims); ok {
			if str, ok := token.Set[claimName].(string); ok && str != "" {
				return str
			}
		}
	}

	return ""
}

func getRoutes(handlerName string) (string, string, string) {
	array := strings.Split(handlerName, ".")
	return array[0], array[1], array[2]
}
