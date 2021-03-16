package shttp

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/syncfuture/go/sconfig"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host/model"
)

func ConfigHttpClient(configProvider sconfig.IConfigProvider) {
	// Http客户端配置
	skipCertVerification := configProvider.GetBool("Http.SkipCertVerification")
	proxy := configProvider.GetString("Http.Proxy")
	if skipCertVerification || proxy != "" {
		// 任意条件满足，则使用自定义传输层
		transport := new(http.Transport)
		if skipCertVerification {
			// 跳过证书验证
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: skipCertVerification}
		}
		if proxy != "" {
			// 使用代理
			transport.Proxy = func(r *http.Request) (*url.URL, error) {
				return url.Parse(proxy)
			}
		}
		http.DefaultClient.Transport = transport
	}
}

func GetUser(ctx IHttpContext, userJsonSessionkey string) (r *model.User) {
	userJson := ctx.GetSessionString(userJsonSessionkey)
	if userJson != "" {
		// 已登录
		err := json.Unmarshal(u.StrToBytes(userJson), &r)
		u.LogError(err)
	}
	return
}

func GetUserID(ctx IHttpContext, userIDSessionkey string) string {
	return ctx.GetSessionString(userIDSessionkey)
}

func SignOut(ctx IHttpContext, tokenCookieName string) {
	ctx.EndSession()
	ctx.RemoveCookie(tokenCookieName)
}
