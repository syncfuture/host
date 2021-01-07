package main

import (
	"sync"

	"github.com/valyala/fasthttp"
)

var _homePagePool = sync.Pool{
	New: func() interface{} {
		return new(homePage)
	},
}

func aquireHomePage() *homePage {
	return _homePagePool.Get().(*homePage)
}
func releaseHomePage(view *homePage) {
	_homePagePool.Put(view)
}

func homePageGet(ctx *fasthttp.RequestCtx) {
	view := aquireHomePage()
	view.Username = _authServer.GetCookie(ctx, _options.AuthCookieName)
	writePage(ctx, view)
	releaseHomePage(view)
}
