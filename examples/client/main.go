package main

import (
	iriscontext "github.com/kataras/iris/v12/context"
	"github.com/syncfuture/host"
)

func main() {
	options := host.NewClientServerOptions()
	s := host.NewClientServer(options)

	var testActions = &[]*host.Action{
		host.NewAction("GET/", "root", "Home", getIndex),
	}

	s.WebServer.Use(s.Authorize)
	s.Run(testActions)
}

func getIndex(ctx iriscontext.Context) {
	ctx.Write([]byte("Home"))
}
