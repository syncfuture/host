package client

import (
	"encoding/json"
	"net/http"
	"time"

	oauth2core "github.com/Lukiya/oauth2go/core"
	"github.com/pascaldekloe/jwt"
	"github.com/syncfuture/go/serr"
	"github.com/syncfuture/go/ssecurity"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host"
	"golang.org/x/oauth2"
)

const _cookieTokenProtectorKey = "token"

type CookieTokenStore struct {
	CookieEncryptor ssecurity.ICookieEncryptor
	TokenCookieName string
}

func NewCookieTokenStore(tokenCookieName string, cookieEncryptor ssecurity.ICookieEncryptor) *CookieTokenStore {
	return &CookieTokenStore{
		TokenCookieName: tokenCookieName,
		CookieEncryptor: cookieEncryptor,
	}
}

/// SaveToken 保存令牌
func (x *CookieTokenStore) SaveToken(ctx host.IHttpContext, token *oauth2.Token) error {
	tokenJson, err := json.Marshal(token)
	if err != nil {
		return serr.WithStack(err)
	}

	// 令牌加密
	securedString, err := x.CookieEncryptor.Encrypt(_cookieTokenProtectorKey, tokenJson)

	// 保存加密后的令牌到Cookie
	tokenCookie := new(http.Cookie)
	tokenCookie.Name = x.TokenCookieName
	tokenCookie.Value = securedString
	tokenCookie.HttpOnly = true
	tokenClaims, err := jwt.ParseWithoutCheck(u.StrToBytes(token.AccessToken))
	if err != nil {
		return serr.WithStack(err)
	}
	if rexp, ok := tokenClaims.Set[oauth2core.Claim_RefreshTokenExpire].(float64); ok {
		// claims里有刷新令牌过期时间，作为Cookie
		tokenCookie.Expires = time.Unix(int64(rexp), 0)
		// 否则作为session存储
	}

	ctx.SetCookieKV(x.TokenCookieName, securedString, func(c *http.Cookie) {
		c.HttpOnly = true
	})
	return nil
}

/// GetToken 获取令牌
func (x *CookieTokenStore) GetToken(ctx host.IHttpContext) (*oauth2.Token, error) {
	// 从Session获取令牌
	tokenJson := ctx.GetCookieString(x.TokenCookieName)
	if tokenJson == "" {
		return nil, nil
	}
	var tokenJsonBytes []byte
	err := x.CookieEncryptor.Decrypt(_cookieTokenProtectorKey, tokenJson, &tokenJsonBytes)
	if err != nil {
		return nil, serr.WithStack(err)
	}

	t := new(oauth2.Token)
	err = json.Unmarshal(tokenJsonBytes, t)
	if err != nil {
		return nil, serr.WithStack(err)
	}

	return t, serr.WithStack(err)
}
