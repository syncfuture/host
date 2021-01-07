package main

import (
	"encoding/json"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/syncfuture/host"
)

var _publicActions = &[]*host.Action{
	host.NewAction("GET/users", "user", "home", getUsers),
}

func main() {
	options := host.NewOAuthResourceOptions()
	options.URIKey = "t:URIS"
	options.RouteKey = "ti:ROUTES"
	options.PermissionKey = "t:PERMISSIONS"
	server := host.NewOAuthResource(options)

	server.Run(_publicActions)
}

func getUsers(ctx iris.Context) {
	a := map[string]string{
		"time": time.Now().String(),
		"abc":  "abc",
		"def":  "def",
		"ghi":  "ghi",
	}
	json, _ := json.Marshal(a)
	ctx.Write(json)
}
