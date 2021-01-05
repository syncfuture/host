package test

import (
	"encoding/json"
	"testing"

	"github.com/kataras/iris/v12"
	"github.com/syncfuture/host"
)

var PublicActions = &[]*host.Action{
	host.NewAction("GET/users", "user", "home", GetUsers),
}

func TestClient(t *testing.T) {
	options := host.NewOAuthResourceOptions()
	options.URIKey = "test:URIS"
	options.RouteKey = "testapi:ROUTES"
	options.PermissionKey = "test:PERMISSIONS"
	api := host.NewOAuthResource(options)

	api.Run(PublicActions)
}

func GetUsers(ctx iris.Context) {
	a := map[string]string{
		"abc": "abc",
		"def": "def",
		"ghi": "ghi",
	}
	json, _ := json.Marshal(a)
	ctx.Write(json)
}
