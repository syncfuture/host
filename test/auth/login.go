package main

import (
	"errors"
	"net/url"
	"strconv"
	"sync"
	"time"

	oauth2 "github.com/Lukiya/oauth2go/core"
	"github.com/valyala/fasthttp"
)

// login
var _loginPagePool = sync.Pool{
	New: func() interface{} {
		return new(loginPage)
	},
}

func aquireLoginPage() *loginPage {
	return _loginPagePool.Get().(*loginPage)
}
func releaseLoginPage(view *loginPage) {
	_loginPagePool.Put(view)
}

func loginPageGet(ctx *fasthttp.RequestCtx) {
	view := aquireLoginPage()
	view.ReturnURL = url.QueryEscape(string(ctx.FormValue(oauth2.Form_ReturnUrl)))
	writePage(ctx, view)
	releaseLoginPage(view)
}

func loginPagePost(ctx *fasthttp.RequestCtx) {
	username := string(ctx.FormValue("Username"))
	password := string(ctx.FormValue("Password"))
	rememberLogin, _ := strconv.ParseBool(string(ctx.FormValue("RememberLogin")))
	returnURL := string(ctx.FormValue(oauth2.Form_ReturnUrl))

	var err error
	if username == password {
		// login success, set login cookie
		if rememberLogin {
			_authServer.SetCookie(ctx, _options.AuthCookieName, username, 24*time.Hour*14)
		} else {
			_authServer.SetCookie(ctx, _options.AuthCookieName, username, 0)
		}
		oauth2.Redirect(ctx, returnURL)
		return
	} else {
		err = errors.New("incorrect password")
	}

	view := aquireLoginPage()
	view.ReturnURL = url.QueryEscape(returnURL)
	view.Error = err.Error()
	writePage(ctx, view)
	releaseLoginPage(view)
}
