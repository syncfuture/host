package sfasthttp

import (
	"embed"
	"mime"
	fp "path/filepath"
	"strings"
	"time"

	"github.com/fasthttp/router"
	"github.com/fasthttp/session/v2"
	"github.com/fasthttp/session/v2/providers/memory"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host/shttp"
	"github.com/valyala/fasthttp"
)

const (
	_filepath = "filepath"
	_suffix   = "/{" + _filepath + ":*}"
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

func (x *FHWebHost) GET(path string, handlers ...shttp.RequestHandler) {
	x.Router.GET(path, AdaptHandler(x.SessionManager, handlers...))
}
func (x *FHWebHost) POST(path string, handlers ...shttp.RequestHandler) {
	x.Router.POST(path, AdaptHandler(x.SessionManager, handlers...))
}
func (x *FHWebHost) PUT(path string, handlers ...shttp.RequestHandler) {
	x.Router.PUT(path, AdaptHandler(x.SessionManager, handlers...))
}
func (x *FHWebHost) DELETE(path string, handlers ...shttp.RequestHandler) {
	x.Router.DELETE(path, AdaptHandler(x.SessionManager, handlers...))
}
func (x *FHWebHost) OPTIONS(path string, handlers ...shttp.RequestHandler) {
	x.Router.OPTIONS(path, AdaptHandler(x.SessionManager, handlers...))
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

func (x *FHWebHost) BuildFHWebHost() {
	if x.SessionCookieName == "" {
		x.SessionCookieName = "go.cookie1"
	}

	////////// router
	if x.Router == nil {
		x.Router = router.New()
		x.Router.PanicHandler = func(ctx *fasthttp.RequestCtx, err interface{}) {
			ctx.SetStatusCode(500)
			log.Error(err)
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
