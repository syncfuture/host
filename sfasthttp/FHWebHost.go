package sfasthttp

import (
	"embed"
	"mime"
	"net/http"
	fp "path/filepath"
	"strings"
	"time"

	"github.com/fasthttp/router"
	"github.com/fasthttp/session/v2"
	"github.com/fasthttp/session/v2/providers/memory"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host/abstracts"
	"github.com/syncfuture/host/shttp"
	"github.com/valyala/fasthttp"
)

const (
	_filepath = "filepath"
	_suffix   = "/{" + _filepath + ":*}"
)

// FHWebHost : IWebHost
type FHWebHost struct {
	base *abstracts.BaseWebHost
	// 独有属性
	SessionCookieName string
	SessionExpSeconds int
	Router            *router.Router
	SessionProvider   session.Provider
	SessionManager    *session.Session
	PanicHandler      shttp.RequestHandler
}

func (x *FHWebHost) BuildFHWebHost(base *abstracts.BaseWebHost) {
	x.base = base
	if x.SessionCookieName == "" {
		x.SessionCookieName = "go.cookie1"
	}

	////////// router
	if x.Router == nil {
		x.Router = router.New()
		x.Router.PanicHandler = func(ctx *fasthttp.RequestCtx, err interface{}) {
			if x.PanicHandler != nil {
				newCtx := NewFastHttpContext(ctx, x.SessionManager)
				newCtx.SetItem(shttp.Item_PANIC, err)
				x.PanicHandler(newCtx)
				return
			}
			ctx.SetStatusCode(500)
			log.Errorf("%s -> %s", ctx.URI().String(), err)
		}
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
		if x.SessionExpSeconds <= 0 {
			cfg.Expiration = -1
		} else {
			cfg.Expiration = time.Second * time.Duration(x.SessionExpSeconds)
		}
		cfg.CookieName = x.SessionCookieName
		cfg.EncodeFunc = session.MSGPEncode // 内存型provider性能较好
		cfg.DecodeFunc = session.MSGPDecode // 内存型provider性能较好

		x.SessionManager = session.New(cfg)
		err := x.SessionManager.SetProvider(x.SessionProvider)
		u.LogFaltal(err)
	}
}

func (x *FHWebHost) GET(path string, handlers ...shttp.RequestHandler) {
	x.Router.GET(path, ToNativeHandler(x.SessionManager, handlers...))
}
func (x *FHWebHost) POST(path string, handlers ...shttp.RequestHandler) {
	x.Router.POST(path, ToNativeHandler(x.SessionManager, handlers...))
}
func (x *FHWebHost) PUT(path string, handlers ...shttp.RequestHandler) {
	x.Router.PUT(path, ToNativeHandler(x.SessionManager, handlers...))
}
func (x *FHWebHost) PATCH(path string, handlers ...shttp.RequestHandler) {
	x.Router.PATCH(path, ToNativeHandler(x.SessionManager, handlers...))
}
func (x *FHWebHost) DELETE(path string, handlers ...shttp.RequestHandler) {
	x.Router.DELETE(path, ToNativeHandler(x.SessionManager, handlers...))
}
func (x *FHWebHost) OPTIONS(path string, handlers ...shttp.RequestHandler) {
	x.Router.OPTIONS(path, ToNativeHandler(x.SessionManager, handlers...))
}

func (x *FHWebHost) ServeFiles(webPath, physiblePath string) {
	x.Router.ServeFiles(webPath, physiblePath)
}

func (x *FHWebHost) ServeEmbedFiles(webPath, physiblePath string, emd embed.FS) {
	if !strings.HasSuffix(webPath, _suffix) {
		panic("path must end with " + _suffix + " in path '" + webPath + "'")
	}

	x.Router.GET(webPath, func(ctx *fasthttp.RequestCtx) {
		filepath := physiblePath + "/" + ctx.UserValue(_filepath).(string)
		file, err := emd.Open(filepath) // embed file doesn't need to close
		if err == nil {
			ext := fp.Ext(filepath)
			cType := mime.TypeByExtension(ext)

			if cType != "" {
				ctx.SetContentType(cType)
			}
			ctx.Response.SetBodyStream(file, -1)
			return
		}

		ctx.SetStatusCode(404)
		ctx.WriteString("NOT FOUND")
	})
}

func (x *FHWebHost) Run(actionGroups ...*abstracts.ActionGroup) error {

	////////// 添加Actions
	x.base.RegisterActionGroups(actionGroups...)

	////////// 注册Actions到路由表
	for k, v := range x.base.Actions {
		x.registerActionToRoute(k, v.Handlers...)
	}

	////////// 开始Serve
	log.Infof("Listening on %s", x.base.ListenAddr)
	return fasthttp.ListenAndServe(x.base.ListenAddr, x.Router.Handler)
}

func (x *FHWebHost) registerActionToRoute(route string, handlers ...shttp.RequestHandler) {
	index := strings.Index(route, "/")
	method := route[:index]
	path := route[index:]

	switch method {
	case http.MethodPost:
		x.Router.POST(path, ToNativeHandler(x.SessionManager, handlers...))
		break
	case http.MethodGet:
		x.Router.GET(path, ToNativeHandler(x.SessionManager, handlers...))
		break
	case http.MethodPut:
		x.Router.PUT(path, ToNativeHandler(x.SessionManager, handlers...))
		break
	case http.MethodPatch:
		x.Router.PATCH(path, ToNativeHandler(x.SessionManager, handlers...))
		break
	case http.MethodDelete:
		x.Router.DELETE(path, ToNativeHandler(x.SessionManager, handlers...))
		break
	case http.MethodOptions:
		x.Router.OPTIONS(path, ToNativeHandler(x.SessionManager, handlers...))
		break
	default:
		panic("does not support method " + method)
	}
}
