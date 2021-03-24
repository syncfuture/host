package host

import "github.com/syncfuture/go/sid"

const (
	// Header_ContentType = "Content-Type"
	Seperator_Route = "_"
	AuthType_Bearer = "Bearer"
	Item_RouteKey   = "RouteKey"
	Item_JWT        = "jwt"
	Item_PANIC      = "panic"
)

var (
	_idGenerator = sid.NewSonyflakeIDGenerator()
)

type (
	RequestHandler func(ctx IHttpContext)
)
