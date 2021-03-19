package host

func JsonConentTypeHandler(ctx IHttpContext) {
	ctx.SetContentType(CType_Json)
	ctx.Next()
}
