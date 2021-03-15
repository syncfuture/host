package sfasthttp

import (
	"github.com/fasthttp/session/v2"
	"github.com/syncfuture/host/shttp"
	"github.com/valyala/fasthttp"
)

func AdaptHandler(handler shttp.RequestHandler, sess *session.Session) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		var newCtx shttp.IHttpContext
		defer PutFastHttpContext(newCtx)

		newCtx = NewFastHttpContext(ctx, sess)
		handler(newCtx)
	})
}
