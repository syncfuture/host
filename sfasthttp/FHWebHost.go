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
	"github.com/syncfuture/go/sconfig"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/ssecurity"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host"
	"github.com/valyala/fasthttp"
)

const (
	_filepath = "filepath"
	_suffix   = "/{" + _filepath + ":*}"
)

type WebHostOption func(*FHWebHost)

// FHWebHost : IWebHost
type FHWebHost struct {
	host.BaseWebHost
	// 独有属性
	SessionCookieName string
	SessionExpSeconds int
	ReadBufferSize    int
	Router            *router.Router
	SessionProvider   session.Provider
	SessionManager    *session.Session
	PanicHandler      host.RequestHandler
	CookieEncryptor   ssecurity.ICookieEncryptor
}

func NewFHWebHost(cp sconfig.IConfigProvider, options ...WebHostOption) host.IWebHost {
	r := new(FHWebHost)
	cp.GetStruct("@this", &r)

	for _, o := range options {
		o(r)
	}

	r.buildFHWebHost()

	return r
}

func (x *FHWebHost) buildFHWebHost() {
	x.BuildBaseWebHost()

	if x.SessionCookieName == "" {
		x.SessionCookieName = "go.cookie1"
	}

	////////// router
	if x.Router == nil {
		x.Router = router.New()
		x.Router.PanicHandler = func(ctx *fasthttp.RequestCtx, err interface{}) {
			if x.PanicHandler != nil {
				newCtx := NewFastHttpContext(ctx, x.SessionManager, x.CookieEncryptor)
				newCtx.SetItem(host.Ctx_Panic, err)
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

	if x.ReadBufferSize <= 0 {
		x.ReadBufferSize = 4096
	}

	////////// CORS
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

func (x *FHWebHost) BuildNativeHandler(routeKey string, handlers ...host.RequestHandler) fasthttp.RequestHandler {
	if len(handlers) == 0 {
		log.Fatal("handlers are missing")
	}

	// 注册全局中间件
	if len(x.GlobalPreHandlers) > 0 {
		handlers = append(x.GlobalPreHandlers, handlers...)
	}
	if len(x.GlobalSufHandlers) > 0 {
		handlers = append(handlers, x.GlobalSufHandlers...)
	}

	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		var newCtx host.IHttpContext
		newCtx = NewFastHttpContext(ctx, x.SessionManager, x.CookieEncryptor, handlers...)
		newCtx.SetItem(host.Ctx_RouteKey, routeKey)
		defer func() {
			newCtx.Reset()
			_ctxPool.Put(newCtx)
		}()
		handlers[0](newCtx) // 开始执行第一个Handler
	})
}

func (x *FHWebHost) GET(path string, handlers ...host.RequestHandler) {
	x.Router.GET(path, x.BuildNativeHandler(path, handlers...))
}
func (x *FHWebHost) POST(path string, handlers ...host.RequestHandler) {
	x.Router.POST(path, x.BuildNativeHandler(path, handlers...))
}
func (x *FHWebHost) PUT(path string, handlers ...host.RequestHandler) {
	x.Router.PUT(path, x.BuildNativeHandler(path, handlers...))
}
func (x *FHWebHost) PATCH(path string, handlers ...host.RequestHandler) {
	x.Router.PATCH(path, x.BuildNativeHandler(path, handlers...))
}
func (x *FHWebHost) DELETE(path string, handlers ...host.RequestHandler) {
	x.Router.DELETE(path, x.BuildNativeHandler(path, handlers...))
}
func (x *FHWebHost) OPTIONS(path string, handlers ...host.RequestHandler) {
	x.Router.OPTIONS(path, x.BuildNativeHandler(path, handlers...))
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

func (x *FHWebHost) Run() error {
	////////// 注册Actions到路由
	for _, v := range x.Actions {
		x.RegisterActionsToRouter(v)
	}

	////////// 开始Serve
	log.Infof("Listening on %s", x.ListenAddr)
	s := &fasthttp.Server{
		Handler:        x.Router.Handler,
		ReadBufferSize: x.ReadBufferSize,
	}
	return s.ListenAndServe(x.ListenAddr)
}

func (x *FHWebHost) RegisterActionsToRouter(action *host.Action) {
	index := strings.Index(action.Route, "/")
	method := action.Route[:index]
	path := action.Route[index:]

	switch method {
	case http.MethodPost:
		x.Router.POST(path, x.BuildNativeHandler(action.RouteKey, action.Handlers...))
		break
	case http.MethodGet:
		x.Router.GET(path, x.BuildNativeHandler(action.RouteKey, action.Handlers...))
		break
	case http.MethodPut:
		x.Router.PUT(path, x.BuildNativeHandler(action.RouteKey, action.Handlers...))
		break
	case http.MethodPatch:
		x.Router.PATCH(path, x.BuildNativeHandler(action.RouteKey, action.Handlers...))
		break
	case http.MethodDelete:
		x.Router.DELETE(path, x.BuildNativeHandler(action.RouteKey, action.Handlers...))
		break
	case http.MethodOptions:
		x.Router.OPTIONS(path, x.BuildNativeHandler(action.RouteKey, action.Handlers...))
		break
	default:
		panic("does not support method " + method)
	}
}
