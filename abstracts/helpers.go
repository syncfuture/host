package abstracts

import (
	"crypto/tls"
	"net/http"
	"net/url"

	"github.com/syncfuture/go/sconfig"
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

// func getRoutes(handlerName string) (string, string, string) {
// 	array := strings.Split(handlerName, ".")
// 	return array[0], array[1], array[2]
// }
