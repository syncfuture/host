package host

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strings"

	"github.com/syncfuture/go/config"
	"github.com/syncfuture/go/security"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/sredis"
	"github.com/syncfuture/go/surl"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
	"github.com/syncfuture/go/u"
)

type (
	Action struct {
		Route      string
		Area       string
		Controller string
		Action     string
		Handler    context.Handler
	}

	ServerBase struct {
		Name                    string
		ListenAddr              string
		ConfigProvider          config.IConfigProvider
		RedisConfig             *sredis.RedisConfig
		URLProvider             surl.IURLProvider
		RoutePermissionProvider security.IRoutePermissionProvider
		PermissionAuditor       security.IPermissionAuditor
		WebServer               *iris.Application
	}
)

func (r *ServerBase) setup(configProvider config.IConfigProvider) {
	// init log
	log.Init(configProvider)

	// config provider
	r.ConfigProvider = configProvider

	// http client
	ConfigHttpClient(r.ConfigProvider)

	// redis config
	r.ConfigProvider.GetStruct("Redis", &r.RedisConfig)
	if r.RedisConfig != nil {
		// URLProvider
		r.URLProvider = surl.NewRedisURLProvider(r.RedisConfig)

		// 权限
		if r.Name != "" {
			r.RoutePermissionProvider = security.NewRedisRoutePermissionProvider(r.Name, r.RedisConfig)
			r.PermissionAuditor = security.NewPermissionAuditor(r.RoutePermissionProvider)
		}
	}

	r.WebServer = iris.New()
	r.WebServer.Logger().SetLevel(log.Level)
	r.WebServer.Use(recover.New())
	r.WebServer.Use(logger.New())
}

func NewAction(route, area, controller string, handler context.Handler) *Action {
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

func HandleError(ctx context.Context, err error) bool {
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
