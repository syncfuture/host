package main

import (
	"encoding/json"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/syncfuture/host/siris"
)

var _publicActions = &[]*siris.Action{
	siris.NewAction("GET/users", "user", "home", getUsers),
}

func main() {
	options := siris.NewOAuthResourceOptions()
	server := siris.NewOAuthResource(options)

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
