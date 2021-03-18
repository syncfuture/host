package sfasthttp

import (
	"testing"
	"time"

	"github.com/fasthttp/session/v2/providers/redis"
	"github.com/syncfuture/go/sconfig"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host/abstracts"
)

func TestWebHost(t *testing.T) {
	cp := sconfig.NewJsonConfigProvider("resource.json")
	host := NewFHWebHost(cp, func(x *FHWebHost) {
		x.GlobalPreHandlers = []abstracts.RequestHandler{func(ctx abstracts.IHttpContext) {
			log.Info("GlobalPreHandlers")
			ctx.Next()
		}}

		x.GlobalSufHandlers = []abstracts.RequestHandler{func(ctx abstracts.IHttpContext) {
			log.Info("GlobalSufHandlers")
			ctx.Next()
		}}
	})

	host.GET("/", func(ctx abstracts.IHttpContext) {
		log.Info("Handler")
		routeKey := ctx.GetItemString(abstracts.Item_RouteKey)
		ctx.WriteString(routeKey)
		ctx.Next()
	})

	log.Fatal(host.Run())
}

func TestClient(t *testing.T) {
	provider, err := redis.New(redis.Config{
		KeyPrefix:   "session",
		Addr:        "127.0.0.1:6379",
		Password:    "Famous901",
		PoolSize:    8,
		IdleTimeout: 30 * time.Second,
	})
	u.LogFaltal(err)

	cp := sconfig.NewJsonConfigProvider("client.json")
	host := NewFHOAuthClientHost(cp, func(x *FHOAuthClientHost) {
		x.SessionProvider = provider
	})

	host.AddAction("GET/", "root__", func(ctx abstracts.IHttpContext) {
		ctx.WriteString("Test")
	})

	log.Fatal(host.Run())
}

func TestResource(t *testing.T) {
	cp := sconfig.NewJsonConfigProvider("resource.json")
	host := NewFHOAuthResourceHost(cp, func(x *FHOAuthResourceHost) {
		x.ConfigProvider = cp
	})

	log.Fatal(host.Run())
}
