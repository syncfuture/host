package host

import (
	"embed"
	"net/http"

	"github.com/syncfuture/go/sconfig"
	"github.com/syncfuture/go/sid"
	"github.com/syncfuture/go/sredis"
	"github.com/syncfuture/go/ssecurity"
	"github.com/syncfuture/go/surl"
	"golang.org/x/oauth2"
)

type (
	IOAuthClientHandler interface {
		SignInHandler(ctx IHttpContext)
		SignInCallbackHandler(ctx IHttpContext)
		SignOutHandler(ctx IHttpContext)
		SignOutCallbackHandler(ctx IHttpContext)
	}

	// IAuthMiddleware interface {
	// 	Serve(next RequestHandler, routes ...string) RequestHandler
	// }

	IBaseHost interface {
		GetConfigProvider() sconfig.IConfigProvider
		GetIDGenerator() sid.IIDGenerator
		GetRedisConfig() *sredis.RedisConfig
		GetURLProvider() surl.IURLProvider
		GetPermissionAuditor() ssecurity.IPermissionAuditor
		GetPermissionProvider() ssecurity.IPermissionProvider
		GetRouteProvider() ssecurity.IRouteProvider
	}

	IWebHost interface {
		GET(path string, handlers ...RequestHandler)
		POST(path string, handlers ...RequestHandler)
		PUT(path string, handlers ...RequestHandler)
		PATCH(path string, handlers ...RequestHandler)
		DELETE(path string, handlers ...RequestHandler)
		OPTIONS(path string, handlers ...RequestHandler)
		ServeFiles(webPath, physiblePath string)
		ServeEmbedFiles(webPath, physiblePath string, emd embed.FS)
		AddActionGroups(actionGroups ...*ActionGroup)
		AddActions(actions ...*Action)
		AddAction(route, routeKey string, handlers ...RequestHandler)
		RegisterActionsToRouter(action *Action)
		Run() error
	}

	IHttpContext interface {
		SetItem(key string, value interface{})
		GetItem(key string) interface{}
		GetItemString(key string) string
		GetItemInt(key string) int
		GetItemInt32(key string) int32
		GetItemInt64(key string) int64

		SetCookie(cookie *http.Cookie)
		GetCookieString(key string) string
		RemoveCookie(key string)

		SetSession(key, value string)
		GetSessionString(key string) string
		RemoveSession(key string)
		EndSession()

		GetFormString(key string) string

		GetBodyString() string
		GetBodyBytes() []byte

		GetParamString(key string) string
		GetParamInt(key string) int
		GetParamInt32(key string) int32
		GetParamInt64(key string) int64

		ReadJSON(objPtr interface{}) error
		ReadQuery(objPtr interface{}) error
		ReadForm(objPtr interface{}) error

		GetHeader(key string) string
		SetHeader(key, value string)

		SetStatusCode(statusCode int)
		SetContentType(cType string)
		WriteString(body string) (int, error)
		WriteBytes(body []byte) (int, error)

		RequestURL() string
		GetRemoteIP() string

		Redirect(url string, statusCode int)
		CopyBodyAndStatusCode(resp *http.Response)

		Next()
		Reset()
	}

	RequestHandler func(ctx IHttpContext)

	IContextTokenStore interface {
		SaveToken(ctx IHttpContext, token *oauth2.Token) error
		GetToken(ctx IHttpContext) (*oauth2.Token, error)
	}
)
