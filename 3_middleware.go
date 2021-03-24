package host

import "github.com/syncfuture/go/shttp"

func JsonConentTypeHandler(ctx IHttpContext) {
	ctx.SetContentType(shttp.CTYPE_JSON)
	ctx.Next()
}
