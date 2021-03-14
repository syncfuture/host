package abstracts

import (
	"github.com/syncfuture/go/sconfig"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/sredis"
	"github.com/syncfuture/go/ssecurity"
	"github.com/syncfuture/go/surl"
	"github.com/syncfuture/host/shttp"
)

type (
	IOAuthClientHandler interface {
		SignInHandler(ctx shttp.IHttpContext)
		SignInCallbackHandler(ctx shttp.IHttpContext)
		SignOutHandler(ctx shttp.IHttpContext)
		SignOutCallbackHandler(ctx shttp.IHttpContext)
	}

	BaseServerOptions struct {
		Debug              bool
		Name               string
		URIKey             string
		RouteKey           string
		PermissionKey      string
		ListenAddr         string
		RedisConfig        *sredis.RedisConfig
		ConfigProvider     sconfig.IConfigProvider
		URLProvider        surl.IURLProvider
		PermissionProvider ssecurity.IPermissionProvider
		RouteProvider      ssecurity.IRouteProvider
		PermissionAuditor  ssecurity.IPermissionAuditor
	}

	BaseServer struct {
		Debug              bool
		Name               string
		URIKey             string
		RouteKey           string
		PermissionKey      string
		ListenAddr         string
		ConfigProvider     sconfig.IConfigProvider
		RedisConfig        *sredis.RedisConfig
		URLProvider        surl.IURLProvider
		PermissionProvider ssecurity.IPermissionProvider
		RouteProvider      ssecurity.IRouteProvider
		PermissionAuditor  ssecurity.IPermissionAuditor
	}
)

func (r *BaseServer) ConfigBaseServer(options *BaseServerOptions) {
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
		options.PermissionProvider = ssecurity.NewRedisPermissionProvider(options.PermissionKey, options.RedisConfig)
	}

	if options.RouteProvider == nil && options.RouteKey != "" && options.RedisConfig != nil {
		options.RouteProvider = ssecurity.NewRedisRouteProvider(options.RouteKey, options.RedisConfig)
	}

	if options.PermissionAuditor == nil && options.PermissionProvider != nil { // RouteProvider 允许为空
		options.PermissionAuditor = ssecurity.NewPermissionAuditor(options.PermissionProvider, options.RouteProvider)
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
