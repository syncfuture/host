package token

import (
	"github.com/Lukiya/oauth2go"
	"github.com/Lukiya/oauth2go/security"
	"github.com/Lukiya/oauth2go/security/rsa"
	"github.com/Lukiya/oauth2go/store/redis"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/srsautil"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host"
)

type IOAuthTokenHost interface {
	host.IBaseHost
	host.IWebHost
	host.ISecureCookieHost
	GetAuthCookieName() string
}

type OAuthTokenHost struct {
	host.BaseHost
	oauth2go.TokenHost
	host.SecureCookieHost
	UserJsonSessionKey string
	UserIDSessionKey   string
	PrivateKeyPath     string
	ClientStoreKey     string
	TokenStoreKey      string
	SecretEncryptor    security.ISecretEncryptor
}

func (x *OAuthTokenHost) BuildOAuthTokenHost() {
	// log.Info(x.SecureCookieHost.GetEncryptedCooke)

	x.BaseHost.BuildBaseHost()

	if x.PrivateKeyPath == "" {
		log.Fatal("missing 'PrivateKeyPath' filed in configuration")
	}
	if x.UserJsonSessionKey == "" {
		x.UserJsonSessionKey = "USERJSON"
	}
	if x.UserIDSessionKey == "" {
		x.UserIDSessionKey = "USERID"
	}
	if x.ClientStoreKey == "" {
		x.ClientStoreKey = "CLIENTS"
	}
	if x.TokenStoreKey == "" {
		x.ClientStoreKey = "t:"
	}

	////////// CookieEncryptor
	if x.CookieEncryptor == nil {
		x.SecureCookieHost.BuildSecureCookieHost()
		x.CookieEncryptor = x.GetCookieEncryptor()
	}

	////////// PrivateKey
	if x.PrivateKey == nil {
		var err error
		x.PrivateKey, err = srsautil.ReadPrivateKeyFromFile(x.PrivateKeyPath)
		u.LogFaltal(err)
	}

	////////// SecretEncryptor
	if x.SecretEncryptor == nil {
		x.SecretEncryptor = rsa.NewRSASecretEncryptor(x.PrivateKeyPath)
	}

	////////// ClientStore
	if x.ClientStore == nil {
		if x.ClientStoreKey == "" {
			log.Fatal("ClientStoreKey cannot be empty")
		}
		if x.RedisConfig == nil {
			log.Fatal("missing 'Redis' section in configuration")
		}
		x.ClientStore = redis.NewRedisClientStore(x.ClientStoreKey, x.SecretEncryptor, x.RedisConfig)
	}
	////////// TokenStore
	if x.TokenStore == nil {
		if x.TokenStoreKey == "" {
			log.Fatal("TokenStoreKey cannot be empty")
		}
		if x.RedisConfig == nil {
			log.Fatal("missing 'Redis' section in configuration")
		}
		x.TokenStore = redis.NewRedisTokenStore(x.TokenStoreKey, x.SecretEncryptor, x.RedisConfig)
	}

	x.TokenHost.BuildTokenHost()
}

func (x *OAuthTokenHost) GetAuthCookieName() string {
	return x.AuthCookieName
}
