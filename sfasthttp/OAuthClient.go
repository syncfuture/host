package sfasthttp

type FHOAuthClient struct {
}

// func NewFHOAuthClient() {
// 	s := new(OAuthClientHandler)
// 	r := router.New()

// 	r.GET("", AdaptHandler(s.SignInHandler))
// }

// func AdaptHandler(handler shttp.RequestHandler) fasthttp.RequestHandler {
// 	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
// 		newCtx := NewFastHttpContext(ctx, nil)
// 		handler(newCtx)
// 	})
// }
