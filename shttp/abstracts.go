package shttp

import (
	"net/http"

	"golang.org/x/oauth2"
)

type IHttpContext interface {
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

	Redirect(url string, statusCode int)
	WriteString(body string) (int, error)
	WriteBytes(body []byte) (int, error)
	SetStatusCode(statusCode int)
	SetContentType(cType string)

	RequestURL() string

	SetHeader(key, value string)
}

type RequestHandler func(ctx IHttpContext)

type IContextTokenStore interface {
	SaveToken(ctx IHttpContext, token *oauth2.Token) error
	GetToken(ctx IHttpContext) (*oauth2.Token, error)
}
