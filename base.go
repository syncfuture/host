package host

import (
	"github.com/syncfuture/go/sconfig"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/sredis"
	"github.com/syncfuture/go/ssecurity"
	"github.com/syncfuture/go/surl"
)

type (
	BaseHostOptions struct {
		Debug              bool
		Name               string
		URIKey             string
		RouteKey           string
		PermissionKey      string
		ListenAddr         string
		RedisConfig        *sredis.RedisConfig `json:"Redis,omitempty"`
		ConfigProvider     sconfig.IConfigProvider
		URLProvider        surl.IURLProvider
		PermissionProvider ssecurity.IPermissionProvider
		RouteProvider      ssecurity.IRouteProvider
		PermissionAuditor  ssecurity.IPermissionAuditor
	}

	HostBase struct {
		BaseHostOptions
	}
)

func (r *HostBase) BuildBaseHost(options *BaseHostOptions) {
	// err := deepcopier.Copy(options).To(r)
	// u.LogFaltal(err)

	if r.Name == "" {
		log.Fatal("Name cannot be empty")
	}
	if r.ListenAddr == "" {
		log.Fatal("ListenAddr cannot be empty")
	}

	if r.ConfigProvider == nil {
		r.ConfigProvider = sconfig.NewJsonConfigProvider()
		log.Fatal("ConfigProvider cannot be nil")
	}

	// if options.RedisConfig == nil {
	// 	options.ConfigProvider.GetStruct("Redis", &options.RedisConfig)
	// 	// if options.RedisConfig == nil {
	// 	// 	log.Fatal("RedisConfig cannot be nil")
	// 	// }
	// }

	if r.URLProvider == nil && r.URIKey != "" && r.RedisConfig != nil {
		r.URLProvider = surl.NewRedisURLProvider(r.URIKey, r.RedisConfig)
	}

	if r.PermissionProvider == nil && r.PermissionKey != "" && r.RedisConfig != nil {
		r.PermissionProvider = ssecurity.NewRedisPermissionProvider(r.PermissionKey, r.RedisConfig)
	}

	if r.RouteProvider == nil && r.RouteKey != "" && r.RedisConfig != nil {
		r.RouteProvider = ssecurity.NewRedisRouteProvider(r.RouteKey, r.RedisConfig)
	}

	if r.PermissionAuditor == nil && r.PermissionProvider != nil { // RouteProvider 允许为空
		r.PermissionAuditor = ssecurity.NewPermissionAuditor(r.PermissionProvider, r.RouteProvider)
	}

	// r.Debug = options.Debug
	// r.Name = options.Name
	// r.URIKey = options.URIKey
	// r.RouteKey = options.RouteKey
	// r.PermissionKey = options.PermissionKey
	// r.ListenAddr = options.ListenAddr
	// r.ConfigProvider = options.ConfigProvider
	// r.RedisConfig = options.RedisConfig
	// r.URLProvider = options.URLProvider
	// r.PermissionProvider = options.PermissionProvider
	// r.RouteProvider = options.RouteProvider
	// r.PermissionAuditor = options.PermissionAuditor

	// log.Init(r.ConfigProvider)
	ConfigHttpClient(r.ConfigProvider)

	return
}
