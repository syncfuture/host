package shttp

import (
	"encoding/json"
	"net/http"
	"time"

	oauth2core "github.com/Lukiya/oauth2go/core"
	"github.com/gorilla/securecookie"
	"github.com/pascaldekloe/jwt"
	"golang.org/x/oauth2"
)

const _cookieTokenProtectorKey = "token"

type ITokenStore interface {
	SaveToken(token *oauth2.Token) error
	GetToken() (*oauth2.Token, error)
}

type CookieTokenStore struct {
	ctx              IHttpContext
	CookieProtoector *securecookie.SecureCookie
	TokenCookieName  string
}

func NewCookieTokenStore(tokenCookieName string, ctx IHttpContext, cookieProtoector *securecookie.SecureCookie) *CookieTokenStore {
	return &CookieTokenStore{
		ctx:              ctx,
		TokenCookieName:  tokenCookieName,
		CookieProtoector: cookieProtoector,
	}
}

/// SaveToken 保存令牌
func (x *CookieTokenStore) SaveToken(token *oauth2.Token) error {
	tokenJson, err := json.Marshal(token)
	if err != nil {
		return err
	}

	// 令牌加密
	securedString, err := x.CookieProtoector.Encode(_cookieTokenProtectorKey, tokenJson)

	// 保存加密后的令牌到Cookie
	tokenCookie := new(http.Cookie)
	tokenCookie.Name = x.TokenCookieName
	tokenCookie.Value = securedString
	tokenCookie.HttpOnly = true
	tokenClaims, err := jwt.ParseWithoutCheck([]byte(token.AccessToken))
	if err != nil {
		return err
	}
	if rexp, ok := tokenClaims.Set[oauth2core.Claim_RefreshTokenExpire].(float64); ok {
		// claims里有刷新令牌过期时间，作为Cookie
		tokenCookie.Expires = time.Unix(int64(rexp), 0)
		// 否则作为session存储
	}

	x.ctx.SetCookie(tokenCookie)
	return nil
}

/// GetToken 获取令牌
func (x *CookieTokenStore) GetToken() (*oauth2.Token, error) {
	// 从Session获取令牌
	tokenJson := x.ctx.GetCookieString(x.TokenCookieName)
	if tokenJson == "" {
		return nil, nil
	}
	var tokenJsonBytes []byte
	err := x.CookieProtoector.Decode(_cookieTokenProtectorKey, tokenJson, &tokenJsonBytes)
	if err != nil {
		return nil, err
	}

	t := new(oauth2.Token)
	err = json.Unmarshal(tokenJsonBytes, t)
	if err != nil {
		return nil, err
	}

	return t, err
}
