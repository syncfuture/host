package main

import (
	"encoding/json"

	iriscontext "github.com/kataras/iris/v12/context"
	"github.com/syncfuture/host"
)

func main() {
	options := host.NewAPIServerOptions()
	s := host.NewAPIServer(options)

	var testActions = &[]*host.Action{
		host.NewAction("GET/posts/new", "post", "Public", getPosts),
	}

	s.Run(testActions)
}

func getPosts(ctx iriscontext.Context) {
	var r = []string{"a", "b", "c"}
	json, err := json.Marshal(r)
	if host.HandleError(ctx, err) {
		return
	}

	ctx.Write(json)
}
