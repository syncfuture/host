package host

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strings"

	"github.com/kataras/iris/v12"
	config "github.com/syncfuture/go/sconfig"
	"github.com/syncfuture/go/u"
)

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

func HandleError(ctx iris.Context, err error) bool {
	if u.LogError(err) {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.WriteString(err.Error())
		return true
	}
	return false
}

func ConfigHttpClient(configProvider config.IConfigProvider) {
	// Http客户端配置
	skipCertVerification := configProvider.GetBool("Http.SkipCertVerification")
	proxy := configProvider.GetString("Http.Proxy")
	if skipCertVerification || proxy != "" {
		// 任意条件满足，则使用自定义传输层
		transport := new(http.Transport)
		if skipCertVerification {
			// 跳过证书验证
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: skipCertVerification}
		}
		if proxy != "" {
			// 使用代理
			transport.Proxy = func(r *http.Request) (*url.URL, error) {
				return url.Parse(proxy)
			}
		}
		http.DefaultClient.Transport = transport
	}
}

func getRoutes(handlerName string) (string, string, string) {
	array := strings.Split(handlerName, ".")
	return array[0], array[1], array[2]
}
