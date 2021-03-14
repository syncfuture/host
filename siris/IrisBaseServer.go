package siris

import (
	"net/http"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
	"github.com/kataras/iris/v12/view"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/host/abstracts"
)

type (
	IrisBaseServer struct {
		abstracts.BaseServer
		ViewsDir       string
		IrisApp        *iris.Application
		ViewEngine     view.Engine
		PreMiddlewares []iris.Handler
		ActionMap      *map[string]*Action
	}
)

func (r *IrisBaseServer) ConfigIrisBaseServer(options *abstracts.IrisBaseServerOptions) {
	r.ConfigBaseServer(&options.BaseServerOptions)
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

func (x *IrisBaseServer) RegisterActions() {
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
