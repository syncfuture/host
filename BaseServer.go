package host

import (
	"net/http"
	"strings"

	"github.com/syncfuture/go/config"
	"github.com/syncfuture/go/security"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/sredis"
	"github.com/syncfuture/go/surl"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
	"github.com/kataras/iris/v12/view"
)

type (
	Action struct {
		Route      string
		Area       string
		Controller string
		Action     string
		Handler    iris.Handler
	}

	BaseServerOptions struct {
		Debug                   bool
		Name                    string
		URIKey                  string
		RouteKey                string
		PermissionKey           string
		ListenAddr              string
		RedisConfig             *sredis.RedisConfig
		ConfigProvider          config.IConfigProvider
		URLProvider             surl.IURLProvider
		RoutePermissionProvider security.IRoutePermissionProvider
		PermissionAuditor       security.IPermissionAuditor
	}

	IrisBaseServerOptions struct {
		BaseServerOptions
		ViewEngine view.Engine
	}

	BaseServer struct {
		Debug                   bool
		Name                    string
		URIKey                  string
		RouteKey                string
		PermissionKey           string
		ListenAddr              string
		ConfigProvider          config.IConfigProvider
		RedisConfig             *sredis.RedisConfig
		URLProvider             surl.IURLProvider
		RoutePermissionProvider security.IRoutePermissionProvider
		PermissionAuditor       security.IPermissionAuditor
	}

	IrisBaseServer struct {
		BaseServer
		IrisApp        *iris.Application
		ViewEngine     view.Engine
		PreMiddlewares []iris.Handler
		ActionMap      *map[string]*Action
	}
)

func (r *BaseServer) configBaseServer(options *BaseServerOptions) {
	if options.Name == "" {
		log.Fatal("Name cannot be empty")
	}
	if options.URIKey == "" {
		log.Fatal("URIKey cannot be empty")
	}
	// if options.RouteKey == "" {
	// 	log.Fatal("RouteKey cannot be empty")
	// }
	if options.PermissionKey == "" {
		log.Fatal("PermissionKey cannot be empty")
	}
	if options.ListenAddr == "" {
		log.Fatal("ListenAddr cannot be empty")
	}

	if options.ConfigProvider == nil {
		log.Fatal("ConfigProvider cannot be nil")
	}

	if options.RedisConfig == nil {
		options.ConfigProvider.GetStruct("Redis", &options.RedisConfig)
		if options.RedisConfig == nil {
			log.Fatal("RedisConfig cannot be nil")
		}
	}

	if options.URLProvider == nil {
		options.URLProvider = surl.NewRedisURLProvider(options.URIKey, options.RedisConfig)
	}

	if options.RoutePermissionProvider == nil {
		options.RoutePermissionProvider = security.NewRedisRoutePermissionProvider(options.RouteKey, options.PermissionKey, options.RedisConfig)
	}

	if options.PermissionAuditor == nil {
		options.PermissionAuditor = security.NewPermissionAuditor(options.RoutePermissionProvider)
	}

	r.Debug = options.Debug
	r.Name = options.Name
	r.URIKey = options.URIKey
	r.RouteKey = options.RouteKey
	r.PermissionKey = options.PermissionKey
	r.ListenAddr = options.ListenAddr
	r.ConfigProvider = options.ConfigProvider
	r.RedisConfig = options.RedisConfig
	r.URLProvider = options.URLProvider
	r.RoutePermissionProvider = options.RoutePermissionProvider
	r.PermissionAuditor = options.PermissionAuditor

	log.Init(r.ConfigProvider)
	ConfigHttpClient(r.ConfigProvider)

	return
}

func (r *IrisBaseServer) configIrisBaseServer(options *IrisBaseServerOptions) {
	r.configBaseServer(&options.BaseServerOptions)

	if options.RouteKey == "" {
		log.Fatal("RouteKey cannot be empty")
	}

	r.IrisApp = iris.New()
	r.IrisApp.Logger().SetLevel(log.Level)
	r.IrisApp.Use(recover.New())
	r.IrisApp.Use(logger.New())

	return
}

func (x *IrisBaseServer) registerActions() {
	for name, action := range *x.ActionMap {
		handlers := append(x.PreMiddlewares, action.Handler)
		x.registerAction(name, handlers...)
	}
}

func (x *IrisBaseServer) registerAction(name string, handlers ...iris.Handler) {
	index := strings.Index(name, "/")
	method := name[:index]
	path := name[index:]

	switch method {
	case http.MethodPost:
		x.IrisApp.Post(path, handlers...)
		break
	case http.MethodGet:
		x.IrisApp.Get(path, handlers...)
		break
	case http.MethodPut:
		x.IrisApp.Put(path, handlers...)
		break
	case http.MethodPatch:
		x.IrisApp.Patch(path, handlers...)
		break
	case http.MethodDelete:
		x.IrisApp.Delete(path, handlers...)
		break
	default:
		panic("does not support method " + method)
	}
}
