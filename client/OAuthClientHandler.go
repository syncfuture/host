package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Lukiya/oauth2go/core"
	oauth2core "github.com/Lukiya/oauth2go/core"
	"github.com/pascaldekloe/jwt"
	"github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/srand"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host"
	"golang.org/x/oauth2"
)

type OAuthClientHandler struct {
	OAuth              *host.OAuthOptions
	ContextTokenStore  host.IContextTokenStore
	UserJsonSessionkey string
	UserIDSessionKey   string
	TokenCookieName    string
}

func NewOAuthClientHandler(
	oauthOptions *host.OAuthOptions,
	contextTokenStore host.IContextTokenStore,
	userJsonSessionkey string,
	userIDSessionKey string,
	tokenCookieName string,
) host.IOAuthClientHandler {
	return &OAuthClientHandler{
		OAuth:              oauthOptions,
		ContextTokenStore:  contextTokenStore,
		UserJsonSessionkey: userJsonSessionkey,
		UserIDSessionKey:   userIDSessionKey,
		TokenCookieName:    tokenCookieName,
	}
}

func (x *OAuthClientHandler) SignInHandler(ctx host.IHttpContext) {
	returnURL := ctx.GetFormString(oauth2core.Form_ReturnUrl)
	if returnURL == "" {
		returnURL = "/"
	}

	userStr := ctx.GetSessionString(x.UserJsonSessionkey)
	if userStr != "" {
		// 已登录
		ctx.Redirect(returnURL, http.StatusFound)
		return
	}

	// 记录请求地址，跳转去登录页面
	host.RedirectAuthorizeEndpoint(ctx, x.OAuth, returnURL)
	// state := srand.String(32)
	// ctx.SetSession(state, returnURL)
	// if x.OAuth.PkceRequired {
	// 	codeVerifier := oauth2core.Random64String()
	// 	codeChanllenge := oauth2core.ToSHA256Base64URL(codeVerifier)
	// 	ctx.SetSession(oauth2core.Form_CodeVerifier, codeVerifier)
	// 	ctx.SetSession(oauth2core.Form_CodeChallengeMethod, oauth2core.Pkce_S256)
	// 	codeChanllengeParam := oauth2.SetAuthURLParam(oauth2core.Form_CodeChallenge, codeChanllenge)
	// 	codeChanllengeMethodParam := oauth2.SetAuthURLParam(oauth2core.Form_CodeChallengeMethod, oauth2core.Pkce_S256)
	// 	ctx.Redirect(x.OAuth.AuthCodeURL(state, codeChanllengeParam, codeChanllengeMethodParam), http.StatusFound)
	// } else {
	// 	ctx.Redirect(x.OAuth.AuthCodeURL(state), http.StatusFound)
	// }
}
func (x *OAuthClientHandler) SignInCallbackHandler(ctx host.IHttpContext) {

	state := ctx.GetFormString(oauth2core.Form_State)
	redirectUrl := ctx.GetSessionString(state)
	if redirectUrl == "" {
		ctx.WriteString("invalid state")
		ctx.SetStatusCode(http.StatusBadRequest)
		return
	}
	ctx.RemoveSession(state) // 释放内存

	var sessionCodeVerifier, sessionSodeChallengeMethod string
	if x.OAuth.PkceRequired {
		sessionCodeVerifier = ctx.GetSessionString(oauth2core.Form_CodeVerifier)
		if sessionCodeVerifier == "" {
			ctx.WriteString("pkce code verifier does not exist in store")
			ctx.SetStatusCode(http.StatusBadRequest)
			return
		}
		ctx.RemoveSession(oauth2core.Form_CodeVerifier)
		sessionSodeChallengeMethod = ctx.GetSessionString(oauth2core.Form_CodeChallengeMethod)
		if sessionCodeVerifier == "" {
			ctx.WriteString("pkce transformation method does not exist in store")
			ctx.SetStatusCode(http.StatusBadRequest)
			return
		}
		ctx.RemoveSession(oauth2core.Form_CodeChallengeMethod)

		codeChallenge := ctx.GetFormString(oauth2core.Form_CodeChallenge)
		codeChallengeMethod := ctx.GetFormString(oauth2core.Form_CodeChallengeMethod)

		if sessionSodeChallengeMethod != codeChallengeMethod {
			ctx.WriteString("pkce transformation method does not match")
			slog.Debugf("session method: '%s', incoming method:'%s'", sessionSodeChallengeMethod, codeChallengeMethod)
			ctx.SetStatusCode(http.StatusBadRequest)
			return
		} else if (sessionSodeChallengeMethod == oauth2core.Pkce_Plain && codeChallenge != oauth2core.ToSHA256Base64URL(sessionCodeVerifier)) ||
			(sessionSodeChallengeMethod == oauth2core.Pkce_Plain && codeChallenge != sessionCodeVerifier) {
			ctx.WriteString("pkce code verifiver and chanllenge does not match")
			slog.Debugf("session verifiver: '%s', incoming chanllenge:'%s'", sessionCodeVerifier, codeChallenge)
			ctx.SetStatusCode(http.StatusBadRequest)
			return
		}
	}

	// 交换令牌
	code := ctx.GetFormString(oauth2core.Form_Code)
	httpCtx := context.Background()
	var oauth2Token *oauth2.Token
	var err error

	// 获取老的刷新令牌，发送给Auth服务器进行注销
	token, _ := x.ContextTokenStore.GetToken(ctx)
	var refreshTokenOption oauth2.AuthCodeOption
	if token != nil && token.RefreshToken != "" {
		refreshTokenOption = oauth2.SetAuthURLParam(oauth2core.Form_RefreshToken, token.RefreshToken)
	}

	if x.OAuth.PkceRequired {
		codeChanllengeParam := oauth2.SetAuthURLParam(oauth2core.Form_CodeVerifier, sessionCodeVerifier)
		codeChanllengeMethodParam := oauth2.SetAuthURLParam(oauth2core.Form_CodeChallengeMethod, sessionSodeChallengeMethod)

		// 发送交换令牌请求
		if refreshTokenOption != nil {
			oauth2Token, err = x.OAuth.Exchange(httpCtx, code, codeChanllengeParam, codeChanllengeMethodParam, refreshTokenOption)
		} else {
			oauth2Token, err = x.OAuth.Exchange(httpCtx, code, codeChanllengeParam, codeChanllengeMethodParam)
		}
	} else {
		if refreshTokenOption != nil {
			oauth2Token, err = x.OAuth.Exchange(httpCtx, code, refreshTokenOption)
		} else {
			oauth2Token, err = x.OAuth.Exchange(httpCtx, code)
		}
	}

	if u.LogError(err) {
		ctx.WriteString(err.Error())
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}

	// 将字符串转化为令牌对象
	jwtToken, err := jwt.ParseWithoutCheck(u.StrToBytes(oauth2Token.AccessToken))
	if err == nil {
		userStr := u.BytesToStr(jwtToken.Raw)
		ctx.SetSession(x.UserJsonSessionkey, userStr)
		if jwtToken.Subject != "" {
			ctx.SetSession(x.UserIDSessionKey, jwtToken.Subject)
		}

		// 保存令牌
		x.ContextTokenStore.SaveToken(ctx, oauth2Token)

		// 重定向到登录前页面
		ctx.Redirect(redirectUrl, http.StatusFound)
	} else {
		ctx.WriteString(err.Error())
		u.LogError(err)
	}
}
func (x *OAuthClientHandler) SignOutHandler(ctx host.IHttpContext) {
	// 去Passport注销
	state := srand.String(32)
	returnUrl := ctx.GetFormString(oauth2core.Form_ReturnUrl)
	if returnUrl == "" {
		returnUrl = "/"
	}
	ctx.SetSession(state, returnUrl)
	targetURL := fmt.Sprintf("%s?%s=%s&%s=%s&%s=%s",
		x.OAuth.EndSessionEndpoint,
		core.Form_ClientID,
		x.OAuth.ClientID,
		core.Form_RedirectUri,
		url.PathEscape(x.OAuth.SignOutRedirectURL),
		core.Form_State,
		url.QueryEscape(state),
	)
	ctx.Redirect(targetURL, http.StatusFound)
}
func (x *OAuthClientHandler) SignOutCallbackHandler(ctx host.IHttpContext) {
	state := ctx.GetFormString(oauth2core.Form_State)
	returnURL := ctx.GetSessionString(state)
	if returnURL == "" {
		ctx.WriteString("invalid state")
		ctx.SetStatusCode(http.StatusBadRequest)
		return
	}

	endSessionID := ctx.GetFormString(oauth2core.Form_EndSessionID)
	if endSessionID == "" {
		ctx.WriteString("missing es_id")
		ctx.SetStatusCode(http.StatusBadRequest)
		return
	}

	token, _ := x.ContextTokenStore.GetToken(ctx)
	if token != nil {
		// 请求Auth服务器删除老RefreshToken
		data := make(url.Values, 5)
		data[oauth2core.Form_State] = []string{state}
		data[oauth2core.Form_EndSessionID] = []string{endSessionID}
		data[oauth2core.Form_ClientID] = []string{x.OAuth.ClientID}
		data[oauth2core.Form_ClientSecret] = []string{x.OAuth.ClientSecret}
		data[oauth2core.Form_RefreshToken] = []string{token.RefreshToken}
		http.PostForm(x.OAuth.EndSessionEndpoint, data)
	}

	host.SignOut(ctx, x.TokenCookieName)

	// 跳转回登出时的页面
	ctx.Redirect(returnURL, http.StatusFound)
}
