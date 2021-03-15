package sfasthttp

import (
	"testing"
	"time"

	"github.com/fasthttp/session/v2/providers/redis"
	"github.com/syncfuture/go/sconfig"
	"github.com/syncfuture/go/u"
)

func TestNewFHOAuthClient(t *testing.T) {
	provider, err := redis.New(redis.Config{
		KeyPrefix:   "session",
		Addr:        "127.0.0.1:6379",
		Password:    "Famous901",
		PoolSize:    8,
		IdleTimeout: 30 * time.Second,
	})
	u.LogFaltal(err)

	cp := sconfig.NewJsonConfigProvider()
	a := NewFHOAuthClientHost(cp, func(x *FHOAuthClientHost) {
		x.ConfigProvider = cp
		x.SessionProvider = provider
	})

	a.Serve()
}
