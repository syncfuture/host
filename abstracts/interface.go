package abstracts

import "github.com/syncfuture/host/shttp"

type (
	IOAuthClientHandler interface {
		SignInHandler(ctx shttp.IHttpContext)
		SignInCallbackHandler(ctx shttp.IHttpContext)
		SignOutHandler(ctx shttp.IHttpContext)
		SignOutCallbackHandler(ctx shttp.IHttpContext)
	}

	IWebHost interface {
		GET(path string, request shttp.RequestHandler)
		POST(path string, request shttp.RequestHandler)
		PUT(path string, request shttp.RequestHandler)
		DELETE(path string, request shttp.RequestHandler)
		Serve() error
	}
)
