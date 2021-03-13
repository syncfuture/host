package shttp

import "net/http"

type IHttpContext interface {
	SetCookie(cookie *http.Cookie)
	GetCookieString(key string) string
	RemoveCookie(key string)

	SetSession(key, value string)
	GetSessionString(key string) string
	RemoveSession(key string)
	EndSession()

	GetFormString(key string) string

	Redirect(url string, statusCode int)
	WriteString(body string) (int, error)
	WriteBytes(body []byte) (int, error)
	SetStatusCode(statusCode int)
	SetContentType(cType string)
}
