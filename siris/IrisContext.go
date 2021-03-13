package siris

import (
	"net/http"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
	"github.com/syncfuture/host/shttp"
)

type IrisContext struct {
	ctx  iris.Context
	sess *sessions.Sessions
}

func NewIrisContext(ctx iris.Context, sess *sessions.Sessions) shttp.IHttpContext {
	return &IrisContext{
		ctx:  ctx,
		sess: sess,
	}
}

func (x *IrisContext) SetCookie(cookie *http.Cookie) {
	x.ctx.SetCookie(cookie)
}
func (x *IrisContext) GetCookieString(key string) string {
	return x.ctx.GetCookie(key)
}
func (x *IrisContext) RemoveCookie(key string) {
	x.ctx.RemoveCookie(key)
}

func (x *IrisContext) SetSession(key, value string) {
	ses := x.sess.Start(x.ctx)
	ses.Set(key, value)
}
func (x *IrisContext) GetSessionString(key string) string {
	ses := x.sess.Start(x.ctx)
	return ses.GetString(key)
}
func (x *IrisContext) RemoveSession(key string) {
	ses := x.sess.Start(x.ctx)
	ses.Delete(key)
}
func (x *IrisContext) EndSession() {
	ses := x.sess.Start(x.ctx)
	ses.Destroy()
}

func (x *IrisContext) GetFormString(key string) string {
	return x.ctx.FormValue(key)
}

func (x *IrisContext) Redirect(url string, statusCode int) {
	x.ctx.Redirect(url, statusCode)
}
func (x *IrisContext) WriteString(body string) (int, error) {
	return x.ctx.WriteString(body)
}
func (x *IrisContext) WriteBytes(body []byte) (int, error) {
	return x.ctx.Write(body)
}
func (x *IrisContext) SetStatusCode(statusCode int) {
	x.ctx.StatusCode(statusCode)
}
func (x *IrisContext) SetContentType(cType string) {
	x.ctx.ContentType(cType)
}
