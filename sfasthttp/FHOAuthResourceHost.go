package sfasthttp

import (
	"github.com/syncfuture/go/sconfig"
	"github.com/syncfuture/host"
	"github.com/syncfuture/host/resource"
)

type ResourceOption func(*FHOAuthResourceHost)

type FHOAuthResourceHost struct {
	resource.OAuthResourceHost
	FHWebHost
}

func NewFHOAuthResourceHost(cp sconfig.IConfigProvider, options ...ResourceOption) resource.IOAuthRespirceHost {
	r := new(FHOAuthResourceHost)
	// r.OAuthResourceHost = new(resource.OAuthResourceHost)
	// r.OAuthResourceHost.BaseHost = new(host.BaseHost)
	// r.FHWebHost = new(FHWebHost)
	cp.GetStruct("@this", &r)
	r.ConfigProvider = cp

	for _, o := range options {
		o(r)
	}

	r.BuildFHOAuthResourceHost()

	return r
}

func (x *FHOAuthResourceHost) BuildFHOAuthResourceHost(options ...ResourceOption) {
	x.BuildOAuthResourceHost()
	x.buildFHWebHost()

	if x.CORS != nil {
		x.AddGlobalPreHandlers(true, func(ctx host.IHttpContext) {
			if x.CORS.AllowedOrigin != "" {
				ctx.SetHeader("Access-Control-Allow-Origin", x.CORS.AllowedOrigin)
			}
			ctx.Next()
		})

		x.OPTIONS("/{filepath:*}", func(ctx host.IHttpContext) {
			// if x.CORS.AllowedOrigin != "" {	// 上面的全局中间件已经添加
			// 	ctx.SetHeader("Access-Control-Allow-Origin", x.CORS.AllowedOrigin)
			// }
			if x.CORS.AllowedMethods != "" {
				ctx.SetHeader("Access-Control-Allow-Methods", x.CORS.AllowedMethods)
			}
			if x.CORS.AllowedHeaders != "" {
				ctx.SetHeader("Access-Control-Allow-Headers", x.CORS.AllowedHeaders)
			}
		})
	}
}
