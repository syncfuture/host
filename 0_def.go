package host

import "github.com/syncfuture/go/sid"

const (
	// Header_ContentType = "Content-Type"
	Seperator_Route = "_"
	AuthType_Bearer = "Bearer"
	Ctx_RouteKey    = "RouteKey"
	Ctx_Claims      = "claims"
	Ctx_Token       = "token"
	Ctx_Panic       = "panic"
)

var (
	_idGenerator = sid.NewSonyflakeIDGenerator()
)

type (
	RequestHandler func(ctx IHttpContext)
)
