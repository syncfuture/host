package sfasthttp

import (
	"time"

	"github.com/fasthttp/router"
	"github.com/fasthttp/session/v2"
	"github.com/fasthttp/session/v2/providers/memory"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host/shttp"
)

// FHWebHost : IWebHost
type FHWebHost struct {
	// 独有属性
	SessionCookieName string
	SessionExpSeconds int
	Router            *router.Router
	SessionProvider   session.Provider
	SessionManager    *session.Session
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

func (x *FHWebHost) BuildFHWebHost() {

	if x.SessionCookieName == "" {
		x.SessionCookieName = "go.cookie1"
	}

	if x.SessionExpSeconds == 0 {
		x.SessionExpSeconds = -1
	}

	////////// router
	if x.Router == nil {
		x.Router = router.New()
	}

	////////// session provider
	if x.SessionProvider == nil {
		provider, err := memory.New(memory.Config{})
		u.LogFaltal(err)
		x.SessionProvider = provider
	}

	////////// session manager
	if x.SessionManager == nil {
		cfg := session.NewDefaultConfig()
		cfg.Expiration = time.Second * time.Duration(x.SessionExpSeconds)
		cfg.CookieName = x.SessionCookieName
		cfg.EncodeFunc = session.MSGPEncode // 内存型provider性能较好
		cfg.DecodeFunc = session.MSGPDecode // 内存型provider性能较好

		x.SessionManager = session.New(cfg)
		err := x.SessionManager.SetProvider(x.SessionProvider)
		u.LogFaltal(err)
	}
}
