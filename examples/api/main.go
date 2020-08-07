package main

import (
	"encoding/json"

	"github.com/kataras/iris/v12/context"
	"github.com/syncfuture/go/config"
	"github.com/syncfuture/host"
)

func main() {
	cp := config.NewJsonConfigProvider()
	var options *host.APIServerOptions
	cp.GetStruct("APIServer", &options)
	s := host.NewAPIServer(cp, options)

	var testActions = &[]*host.Action{
		host.NewAction("GET/posts/new", "post", "Public", getPosts),
	}

	s.Init(testActions)
	s.Run()
}

func getPosts(ctx context.Context) {
	var r = []string{"a", "b", "c"}
	json, err := json.Marshal(r)
	if host.HandleError(ctx, err) {
		return
	}

	ctx.Write(json)
}
