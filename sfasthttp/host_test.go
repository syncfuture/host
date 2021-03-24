package sfasthttp

import (
	"testing"
	"time"

	"github.com/Lukiya/oauth2go/model"
	"github.com/fasthttp/session/v2/providers/redis"
	"github.com/syncfuture/go/sconfig"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host"
)

func TestWebHost(t *testing.T) {
	cp := sconfig.NewJsonConfigProvider("resource.json")
	h := NewFHWebHost(cp, func(x *FHWebHost) {
		x.GlobalPreHandlers = []host.RequestHandler{func(ctx host.IHttpContext) {
			log.Info("GlobalPreHandlers")
			ctx.Next()
		}}

		x.GlobalSufHandlers = []host.RequestHandler{func(ctx host.IHttpContext) {
			log.Info("GlobalSufHandlers")
			ctx.Next()
		}}
	})

	h.GET("/", func(ctx host.IHttpContext) {
		log.Info("Handler")
		routeKey := ctx.GetItemString(host.Item_RouteKey)
		ctx.WriteString(routeKey)
		ctx.Next()
	})

	log.Fatal(h.Run())
}

func TestClientHost(t *testing.T) {
	provider, err := redis.New(redis.Config{
		KeyPrefix:   "session",
		Addr:        "127.0.0.1:6379",
		Password:    "Famous901",
		PoolSize:    8,
		IdleTimeout: 30 * time.Second,
	})
	u.LogFaltal(err)

	cp := sconfig.NewJsonConfigProvider("client.json")
	h := NewFHOAuthClientHost(cp, func(x *FHOAuthClientHost) {
		x.SessionProvider = provider
	})

	h.AddAction("GET/", "root__", func(ctx host.IHttpContext) {
		ctx.WriteString("Test")
	})

	log.Fatal(h.Run())
}

func TestResourceHost(t *testing.T) {
	cp := sconfig.NewJsonConfigProvider("resource.json")
	h := NewFHOAuthResourceHost(cp)

	log.Fatal(h.Run())
}

func TestTokenHost(t *testing.T) {
	cp := sconfig.NewJsonConfigProvider("token.json")
	h := NewFHOAuthTokenHost(cp, func(x *FHOAuthTokenHost) {
		x.ClaimsGenerator = &testClaimsGenerator{}
		// x.ResourceOwnerValidator = nil
	})

	log.Fatal(h.Run())
}

type testClaimsGenerator struct{}

func (x *testClaimsGenerator) Generate(grantType string, client model.IClient, scopes []string, username string) *map[string]interface{} {
	return &map[string]interface{}{}
}
