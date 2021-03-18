package host

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	oauth2core "github.com/Lukiya/oauth2go/core"
	"github.com/pascaldekloe/jwt"
	"github.com/syncfuture/go/sconfig"
	"github.com/syncfuture/go/sconv"
	"github.com/syncfuture/go/srand"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host/model"
	"golang.org/x/oauth2"
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

func GetRoutesByKey(routeKey string) (area, controller, action string) {
	routeArray := strings.Split(routeKey, Seperator_Route)

	if len(routeArray) >= 3 {
		action = routeArray[2]
	}

	if len(routeArray) >= 2 {
		controller = routeArray[1]
	}

	area = routeArray[0]

	return
}

func GetUserID(ctx IHttpContext, userIDSessionkey string) string {
	return ctx.GetSessionString(userIDSessionkey)
}

func SignOut(ctx IHttpContext, tokenCookieName string) {
	ctx.EndSession()
	ctx.RemoveCookie(tokenCookieName)
}

func RedirectAuthorizeEndpoint(ctx IHttpContext, oauthOptions *OAuthOptions, returnURL string) {
	state := srand.String(32)
	ctx.SetSession(state, returnURL)
	if oauthOptions.PkceRequired {
		codeVerifier := oauth2core.Random64String()
		codeChanllenge := oauth2core.ToSHA256Base64URL(codeVerifier)
		ctx.SetSession(oauth2core.Form_CodeVerifier, codeVerifier)
		ctx.SetSession(oauth2core.Form_CodeChallengeMethod, oauth2core.Pkce_S256)
		codeChanllengeParam := oauth2.SetAuthURLParam(oauth2core.Form_CodeChallenge, codeChanllenge)
		codeChanllengeMethodParam := oauth2.SetAuthURLParam(oauth2core.Form_CodeChallengeMethod, oauth2core.Pkce_S256)
		ctx.Redirect(oauthOptions.AuthCodeURL(state, codeChanllengeParam, codeChanllengeMethodParam), http.StatusFound)
	} else {
		ctx.Redirect(oauthOptions.AuthCodeURL(state), http.StatusFound)
	}
}

func GetClaims(ctx IHttpContext) *jwt.Claims {
	j, ok := ctx.GetItem(Item_JWT).(*jwt.Claims)
	if ok {
		return j
	}
	return nil
}

func GetClaimValue(ctx IHttpContext, claimName string) interface{} {
	j := GetClaims(ctx)
	if j != nil {
		if v, ok := j.Set[claimName]; ok {
			return v
		}
	}
	return nil
}

func GetClaimString(ctx IHttpContext, claimName string) string {
	v := GetClaimValue(ctx, claimName)
	return sconv.ToString(v)
}

func GetClaimInt64(ctx IHttpContext, claimName string) int64 {
	v := GetClaimValue(ctx, claimName)
	return sconv.ToInt64(v)
}
