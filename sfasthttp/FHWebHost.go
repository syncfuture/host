package sfasthttp

import (
	"github.com/fasthttp/router"
	"github.com/fasthttp/session/v2"
	"github.com/syncfuture/host/shttp"
)

// FHWebHost : IWebHost
type FHWebHost struct {
	// 独有属性
	Router          *router.Router
	SessionProvider session.Provider
	SessionManager  *session.Session
}

func (x *FHWebHost) GET(path string, request shttp.RequestHandler) {
	x.Router.GET(path, AdaptHandler(request, x.SessionManager))
}
func (x *FHWebHost) POST(path string, request shttp.RequestHandler) {
	x.Router.POST(path, AdaptHandler(request, x.SessionManager))
}
func (x *FHWebHost) PUT(path string, request shttp.RequestHandler) {
	x.Router.PUT(path, AdaptHandler(request, x.SessionManager))
}
func (x *FHWebHost) DELETE(path string, request shttp.RequestHandler) {
	x.Router.DELETE(path, AdaptHandler(request, x.SessionManager))
}
