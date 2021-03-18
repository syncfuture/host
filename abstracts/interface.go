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
		PATCH(path string, request shttp.RequestHandler)
		DELETE(path string, request shttp.RequestHandler)
		OPTIONS(path string, request shttp.RequestHandler)
		Run(actionGroups ...*ActionGroup) error
	}

	IAuthMiddleware interface {
		Serve(next shttp.RequestHandler, routes ...string) shttp.RequestHandler
	}
)
