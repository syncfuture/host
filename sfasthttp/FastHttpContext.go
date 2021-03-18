package sfasthttp

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/fasthttp/session/v2"
	"github.com/iris-contrib/schema"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host/shttp"
	"github.com/valyala/fasthttp"
)

var (
	_ctxPool = &sync.Pool{
		New: func() interface{} {
			return new(FastHttpContext)
		},
	}
)

type FastHttpContext struct {
	ctx          *fasthttp.RequestCtx
	sess         *session.Session
	sessStore    *session.Store
	mapPool      *sync.Pool
	strSlicePool *sync.Pool
	decoder      *schema.Decoder
}

func NewFastHttpContext(ctx *fasthttp.RequestCtx, sess *session.Session) shttp.IHttpContext {
	r := _ctxPool.Get().(*FastHttpContext)
	r.ctx = ctx
	r.sess = sess
	var err error
	r.sessStore, err = r.sess.Get(ctx)
	u.LogFaltal(err)
	r.mapPool = &sync.Pool{
		New: func() interface{} {
			return make(map[string][]string)
		},
	}
	r.strSlicePool = &sync.Pool{
		New: func() interface{} {
			return make([]string, 1)
		},
	}
	r.decoder = schema.NewDecoder()
	return r
}

func PutFastHttpContext(ctx shttp.IHttpContext) {
	_ctxPool.Put(ctx)
}

func (x *FastHttpContext) GetItem(key string) interface{} {
	return x.ctx.UserValue(key)
}
func (x *FastHttpContext) SetItem(key string, value interface{}) {
	x.ctx.SetUserValue(key, value)
}

func (x *FastHttpContext) SetCookie(cookie *http.Cookie) {
	c := new(fasthttp.Cookie)
	c.SetKey(cookie.Name)
	c.SetValue(cookie.Value)
	c.SetMaxAge(cookie.MaxAge)
	c.SetDomain(cookie.Domain)
	c.SetPath(cookie.Path)
	c.SetSecure(cookie.Secure)
	c.SetHTTPOnly(cookie.HttpOnly)
	c.SetExpire(cookie.Expires)
	x.ctx.Response.Header.SetCookie(c)
}
func (x *FastHttpContext) GetCookieString(key string) string {
	r := x.ctx.Request.Header.Cookie(key)
	return u.BytesToStr(r)
}
func (x *FastHttpContext) RemoveCookie(key string) {
	x.ctx.Response.Header.DelClientCookie(key)
}

func (x *FastHttpContext) SetSession(key, value string) {
	store, err := x.sess.Get(x.ctx)
	if u.LogError(err) {
		return
	}
	defer func() {
		u.LogError(x.sess.Save(x.ctx, store))
	}()
	store.Set(key, value)
}
func (x *FastHttpContext) GetSessionString(key string) string {
	store, err := x.sess.Get(x.ctx)
	if u.LogError(err) {
		return ""
	}
	defer func() {
		u.LogError(x.sess.Save(x.ctx, store))
	}()

	if r, ok := store.Get(key).(string); ok {
		return r
	}

	return ""
}
func (x *FastHttpContext) RemoveSession(key string) {
	store, err := x.sess.Get(x.ctx)
	if u.LogError(err) {
		return
	}
	defer func() {
		u.LogError(x.sess.Save(x.ctx, store))
	}()
	store.Delete(key)
}
func (x *FastHttpContext) EndSession() {
	x.sess.Destroy(x.ctx)
	// store, err := x.sess.Get(x.ctx)
	// if u.LogError(err) {
	// 	return
	// }

	// store.Reset()
}

func (x *FastHttpContext) GetFormString(key string) string {
	r := x.ctx.FormValue(key)
	return u.BytesToStr(r)
}

func (x *FastHttpContext) GetBodyString() string {
	return x.ctx.Request.String()
}
func (x *FastHttpContext) GetBodyBytes() []byte {
	return x.ctx.Request.Body()
}

func (x *FastHttpContext) GetParamString(key string) string {
	r, _ := x.ctx.UserValue(key).(string)
	return r
}
func (x *FastHttpContext) GetParamIntDefault(key string, def int) int {
	r, ok := x.ctx.UserValue(key).(int)
	if ok {
		return r
	}
	return def
}
func (x *FastHttpContext) GetParamInt32Default(key string, def int32) int32 {
	r, ok := x.ctx.UserValue(key).(int32)
	if ok {
		return r
	}
	return def
}

func (x *FastHttpContext) ReadJSON(objPtr interface{}) error {
	data := x.ctx.Request.Body()
	return json.Unmarshal(data, objPtr)
}

func (x *FastHttpContext) ReadQuery(objPtr interface{}) error {
	dic := x.mapPool.Get().(map[string][]string)
	defer func() {
		x.mapPool.Put(dic)
	}()
	x.ctx.QueryArgs().VisitAll(func(key, value []byte) {
		v := x.strSlicePool.Get().([]string)
		v[0] = u.BytesToStr(value)
		dic[u.BytesToStr(key)] = v
		x.strSlicePool.Put(v)
	})

	return x.decoder.Decode(objPtr, dic)
}

func (x *FastHttpContext) ReadForm(objPtr interface{}) error {
	dic := x.mapPool.Get().(map[string][]string)
	defer func() {
		x.mapPool.Put(dic)
	}()
	x.ctx.PostArgs().VisitAll(func(key, value []byte) {
		v := x.strSlicePool.Get().([]string)
		v[0] = u.BytesToStr(value)
		dic[u.BytesToStr(key)] = v
		x.strSlicePool.Put(v)
	})

	return x.decoder.Decode(objPtr, dic)
}

func (x *FastHttpContext) Redirect(url string, statusCode int) {
	x.ctx.Redirect(url, statusCode)
}
func (x *FastHttpContext) WriteString(body string) (int, error) {
	return x.ctx.WriteString(body)
}
func (x *FastHttpContext) WriteBytes(body []byte) (int, error) {
	return x.ctx.Write(body)
}

func (x *FastHttpContext) CopyBodyAndStatusCode(resp *http.Response) {
	x.ctx.SetStatusCode(resp.StatusCode)
	x.ctx.SetBodyStream(resp.Body, -1)
}
func (x *FastHttpContext) SetStatusCode(statusCode int) {
	x.ctx.SetStatusCode(statusCode)
}
func (x *FastHttpContext) SetContentType(cType string) {
	x.ctx.SetContentType(cType)
}

func (x *FastHttpContext) RequestURL() string {
	return x.ctx.URI().String()
}

func (x *FastHttpContext) SetHeader(key, value string) {
	x.ctx.Response.Header.Set(key, value)
}

func (x *FastHttpContext) GetHeader(key string) string {
	v := x.ctx.Request.Header.Peek(key)
	return u.BytesToStr(v)
}

func (x *FastHttpContext) GetRemoteIP() string {
	return x.ctx.RemoteIP().String()
}
