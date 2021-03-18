package sfasthttp

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/fasthttp/session/v2"
	"github.com/gorilla/schema"
	"github.com/syncfuture/go/sconv"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host/abstracts"
	"github.com/valyala/fasthttp"
)

var (
	_ctxPool = &sync.Pool{
		New: func() interface{} {
			return new(FastHttpContext)
		},
	}
	_decoder = schema.NewDecoder()
)

type FastHttpContext struct {
	ctx          *fasthttp.RequestCtx
	sess         *session.Session
	sessStore    *session.Store
	mapPool      *sync.Pool
	handlers     []abstracts.RequestHandler
	handlerIndex int
	handlerCount int
}

func NewFastHttpContext(ctx *fasthttp.RequestCtx, sess *session.Session, handlers ...abstracts.RequestHandler) abstracts.IHttpContext {
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
	r.handlers = handlers
	r.handlerCount = len(handlers)
	return r
}

func (x *FastHttpContext) SetItem(key string, value interface{}) {
	x.ctx.SetUserValue(key, value)
}
func (x *FastHttpContext) GetItem(key string) interface{} {
	return x.ctx.UserValue(key)
}
func (x *FastHttpContext) GetItemString(key string) string {
	v := x.ctx.UserValue(key)
	return sconv.ToString(v)
}
func (x *FastHttpContext) GetItemInt(key string) int {
	v := x.ctx.UserValue(key)
	return sconv.ToInt(v)
}
func (x *FastHttpContext) GetItemInt32(key string) int32 {
	v := x.ctx.UserValue(key)
	return sconv.ToInt32(v)
}
func (x *FastHttpContext) GetItemInt64(key string) int64 {
	v := x.ctx.UserValue(key)
	return sconv.ToInt64(v)
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
	v := x.ctx.UserValue(key)
	return sconv.ToString(v)
}
func (x *FastHttpContext) GetParamInt(key string) int {
	v := x.ctx.UserValue(key)
	return sconv.ToInt(v)
}
func (x *FastHttpContext) GetParamInt32(key string) int32 {
	v := x.ctx.UserValue(key)
	return sconv.ToInt32(v)
}
func (x *FastHttpContext) GetParamInt64(key string) int64 {
	v := x.ctx.UserValue(key)
	return sconv.ToInt64(v)
}

func (x *FastHttpContext) ReadJSON(objPtr interface{}) error {
	data := x.ctx.Request.Body()
	return json.Unmarshal(data, objPtr)
}
func (x *FastHttpContext) ReadQuery(objPtr interface{}) error {
	dic := x.mapPool.Get().(map[string][]string)
	defer func() {
		for k := range dic { // this will compile to use "mapclear" internal function
			delete(dic, k)
		}
		x.mapPool.Put(dic)
	}()
	x.ctx.QueryArgs().VisitAll(func(key, value []byte) {
		dic[u.BytesToStr(key)] = []string{u.BytesToStr(value)}
	})

	return _decoder.Decode(objPtr, dic)
}
func (x *FastHttpContext) ReadForm(objPtr interface{}) error {
	dic := x.mapPool.Get().(map[string][]string)
	defer func() {
		for k := range dic { // this will compile to use "mapclear" internal function
			delete(dic, k)
		}
		x.mapPool.Put(dic)
	}()
	x.ctx.PostArgs().VisitAll(func(key, value []byte) {
		dic[u.BytesToStr(key)] = []string{u.BytesToStr(value)}
	})

	return _decoder.Decode(objPtr, dic)
}

func (x *FastHttpContext) SetHeader(key, value string) {
	x.ctx.Response.Header.Set(key, value)
}
func (x *FastHttpContext) GetHeader(key string) string {
	v := x.ctx.Request.Header.Peek(key)
	return u.BytesToStr(v)
}

func (x *FastHttpContext) SetStatusCode(statusCode int) {
	x.ctx.SetStatusCode(statusCode)
}
func (x *FastHttpContext) SetContentType(cType string) {
	x.ctx.SetContentType(cType)
}
func (x *FastHttpContext) WriteString(body string) (int, error) {
	return x.ctx.WriteString(body)
}
func (x *FastHttpContext) WriteBytes(body []byte) (int, error) {
	return x.ctx.Write(body)
}

func (x *FastHttpContext) RequestURL() string {
	return x.ctx.URI().String()
}
func (x *FastHttpContext) GetRemoteIP() string {
	return x.ctx.RemoteIP().String()
}

func (x *FastHttpContext) Redirect(url string, statusCode int) {
	x.ctx.Redirect(url, statusCode)
}
func (x *FastHttpContext) CopyBodyAndStatusCode(resp *http.Response) {
	x.ctx.SetStatusCode(resp.StatusCode)
	x.ctx.SetBodyStream(resp.Body, -1)
}

func (x *FastHttpContext) Next() {
	if x.handlers == nil {
		return
	}

	if x.handlerIndex < x.handlerCount-1 {
		x.handlerIndex++
		x.handlers[x.handlerIndex](x)
	}
}
func (x *FastHttpContext) Reset() {
	x.ctx = nil
	x.sess = nil
	x.sessStore = nil
	x.mapPool = nil
	x.handlers = nil
	x.handlerCount = 0
	x.handlerIndex = 0
}
