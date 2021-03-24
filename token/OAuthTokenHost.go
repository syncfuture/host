package token

import (
	stdrsa "crypto/rsa"

	"github.com/Lukiya/oauth2go/security"
	"github.com/Lukiya/oauth2go/security/rsa"
	"github.com/Lukiya/oauth2go/store"
	"github.com/Lukiya/oauth2go/store/redis"
	"github.com/gorilla/securecookie"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/srsautil"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host"
)

type IOAuthTokenHost interface {
	host.IBaseHost
	host.IWebHost
}

type OAuthTokenHost struct {
	host.BaseHost
	OAuthOptions       *host.OAuthOptions `json:"OAuth,omitempty"`
	HashKey            string
	BlockKey           string
	UserJsonSessionKey string
	UserIDSessionKey   string
	PrivateKeyPath     string
	ClientStoreKey     string
	TokenStoreKey      string
	CookieProtoector   *securecookie.SecureCookie
	ClientStore        store.IClientStore
	TokenStore         store.ITokenStore
	PrivateKey         *stdrsa.PrivateKey
	SecretEncryptor    security.ISecretEncryptor
}

func (x *OAuthTokenHost) BuildOAuthTokenHost() {
	// if x.BaseWebHost == nil {
	// 	x.BaseWebHost = new(host.BaseWebHost)
	// }
	x.BaseHost.BuildBaseHost()

	if x.OAuthOptions == nil {
		log.Fatal("OAuth secion in configuration is missing")
	}
	x.OAuthOptions.BuildOAuthOptions(x.URLProvider)

	if x.BlockKey == "" {
		log.Fatal("block key cannot be empty")
	}
	if x.HashKey == "" {
		log.Fatal("hash key cannot be empty")
	}
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

	////////// CookieProtoector
	if x.CookieProtoector == nil {
		x.CookieProtoector = securecookie.New(u.StrToBytes(x.HashKey), u.StrToBytes(x.BlockKey))
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

	////////// SecretEncryptor
	if x.ClientStore == nil {
		if x.ClientStoreKey == "" {
			log.Fatal("ClientStoreKey cannot be empty")
		}
		if x.RedisConfig == nil {
			log.Fatal("missing 'Redis' section in configuration")
		}
		x.ClientStore = redis.NewRedisClientStore(x.ClientStoreKey, x.SecretEncryptor, x.RedisConfig)
	}
	////////// SecretEncryptor
	if x.TokenStore == nil {
		if x.TokenStoreKey == "" {
			log.Fatal("TokenStoreKey cannot be empty")
		}
		if x.RedisConfig == nil {
			log.Fatal("missing 'Redis' section in configuration")
		}
		x.TokenStore = redis.NewRedisTokenStore(x.TokenStoreKey, x.SecretEncryptor, x.RedisConfig)
	}

	// oauth2go.NewDefaultAuthServer(&oauth2go.AuthServerOptios)
}
