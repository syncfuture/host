package shttp

import (
	"net/http"

	"golang.org/x/oauth2"
)

const (
	Header_Auth     = "Authorization"
	AuthType_Bearer = "Bearer"
	Item_JWT        = "jwt"
)

type IHttpContext interface {
	GetItem(key string) interface{}
	SetItem(key string, value interface{})

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
	ReadJSON(objPtr interface{}) error

	Redirect(url string, statusCode int)
	WriteString(body string) (int, error)
	WriteBytes(body []byte) (int, error)
	CopyBodyAndStatusCode(resp *http.Response)
	SetStatusCode(statusCode int)
	SetContentType(cType string)

	RequestURL() string

	GetHeader(key string) string
	SetHeader(key, value string)
	GetRemoteIP() string
}

type RequestHandler func(ctx IHttpContext)

type IContextTokenStore interface {
	SaveToken(ctx IHttpContext, token *oauth2.Token) error
	GetToken(ctx IHttpContext) (*oauth2.Token, error)
}
