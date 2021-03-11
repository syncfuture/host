package host

import (
	"net/http"
	"strings"

	config "github.com/syncfuture/go/sconfig"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/sredis"
	security "github.com/syncfuture/go/ssecurity"
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
		Debug              bool
		Name               string
		URIKey             string
		RouteKey           string
		PermissionKey      string
		ListenAddr         string
		RedisConfig        *sredis.RedisConfig
		ConfigProvider     config.IConfigProvider
		URLProvider        surl.IURLProvider
		PermissionProvider security.IPermissionProvider
		RouteProvider      security.IRouteProvider
		PermissionAuditor  security.IPermissionAuditor
	}

	IrisBaseServerOptions struct {
		BaseServerOptions
		ViewEngine view.Engine
		ViewsDir   string
	}

	BaseServer struct {
		Debug              bool
		Name               string
		URIKey             string
		RouteKey           string
		PermissionKey      string
		ListenAddr         string
		ConfigProvider     config.IConfigProvider
		RedisConfig        *sredis.RedisConfig
		URLProvider        surl.IURLProvider
		PermissionProvider security.IPermissionProvider
		RouteProvider      security.IRouteProvider
		PermissionAuditor  security.IPermissionAuditor
	}

	IrisBaseServer struct {
		BaseServer
		ViewsDir       string
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
	// if options.URIKey == "" {
	// 	log.Fatal("URIKey cannot be empty")
	// }
	// if options.RouteKey == "" {
	// 	log.Fatal("RouteKey cannot be empty")
	// }
	// if options.PermissionKey == "" {
	// 	log.Fatal("PermissionKey cannot be empty")
	// }
	if options.ListenAddr == "" {
		log.Fatal("ListenAddr cannot be empty")
	}

	if options.ConfigProvider == nil {
		log.Fatal("ConfigProvider cannot be nil")
	}

	if options.RedisConfig == nil {
		options.ConfigProvider.GetStruct("Redis", &options.RedisConfig)
		// if options.RedisConfig == nil {
		// 	log.Fatal("RedisConfig cannot be nil")
		// }
	}

	if options.URLProvider == nil && options.URIKey != "" && options.RedisConfig != nil {
		options.URLProvider = surl.NewRedisURLProvider(options.URIKey, options.RedisConfig)
	}

	if options.PermissionProvider == nil && options.PermissionKey != "" && options.RedisConfig != nil {
		options.PermissionProvider = security.NewRedisPermissionProvider(options.PermissionKey, options.RedisConfig)
	}

	if options.RouteProvider == nil && options.RouteKey != "" && options.RedisConfig != nil {
		options.RouteProvider = security.NewRedisRouteProvider(options.RouteKey, options.RedisConfig)
	}

	if options.PermissionAuditor == nil && options.PermissionProvider != nil { // RouteProvider 允许为空
		options.PermissionAuditor = security.NewPermissionAuditor(options.PermissionProvider, options.RouteProvider)
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
	r.PermissionProvider = options.PermissionProvider
	r.RouteProvider = options.RouteProvider
	r.PermissionAuditor = options.PermissionAuditor

	log.Init(r.ConfigProvider)
	ConfigHttpClient(r.ConfigProvider)

	return
}

func (r *IrisBaseServer) configIrisBaseServer(options *IrisBaseServerOptions) {
	r.configBaseServer(&options.BaseServerOptions)
	if options.ViewsDir == "" {
		options.ViewsDir = "./views"
	}
	r.ViewsDir = options.ViewsDir

	// 创建Iris App
	r.IrisApp = iris.New()
	r.IrisApp.Logger().SetLevel(log.Config.Level)
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
