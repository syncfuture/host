package host

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	oauth2core "github.com/Lukiya/oauth2go/core"
	"github.com/syncfuture/go/sconfig"
	"github.com/syncfuture/go/sconv"
	"github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/srand"
	"github.com/syncfuture/go/ssecurity"
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

func GenerateID() string {
	return _idGenerator.GenerateString()
}

func HandleErr(err error, ctx IHttpContext) bool {
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		errID := _idGenerator.GenerateString()
		slog.Errorf("[%s] %+v", errID, err)
		ctx.WriteJsonBytes(u.StrToBytes(`{"err":"` + errID + `"}`))

		return true
	}
	return false
}

func getClaims(ctx IHttpContext) *map[string]interface{} {
	j, ok := ctx.GetItem(Ctx_Claims).(*map[string]interface{}) // RL00001

	if ok {
		return j
	}

	return nil
}

func getClaimValue(ctx IHttpContext, claimName string) interface{} {
	claims := getClaims(ctx)
	if claims != nil {
		if v, ok := (*claims)[claimName]; ok {
			return v
		}
	}
	return nil
}

func GetClaimString(ctx IHttpContext, claimName string) string {
	v := getClaimValue(ctx, claimName)
	return sconv.ToString(v)
}

func GetClaimInt64(ctx IHttpContext, claimName string) int64 {
	v := getClaimValue(ctx, claimName)
	return sconv.ToInt64(v)
}

func GetEncryptedCookie(ctx IHttpContext, cookieEncryptor ssecurity.ICookieEncryptor, name string) string {
	encryptedCookie := ctx.GetCookieString(name)
	if encryptedCookie == "" {
		return ""
	}

	var r string
	err := cookieEncryptor.Decrypt(name, encryptedCookie, &r)

	if u.LogError(err) {
		return ""
	}

	return r
}
func SetEncryptedCookie(ctx IHttpContext, cookieEncryptor ssecurity.ICookieEncryptor, key, value string, options ...func(*http.Cookie)) {
	if encryptedCookie, err := cookieEncryptor.Encrypt(key, value); err == nil {
		ctx.SetCookieKV(key, encryptedCookie, options...)
	} else {
		u.LogError(err)
	}
}
