package main

import (
	"bytes"
	"net/http"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/kataras/iris/v12/sessions"
	"github.com/syncfuture/go/shttp"
	"github.com/syncfuture/go/surl"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host"
)

var (
	_server *host.OAuthClient
)

func main() {
	options := host.NewOAuthClientOptions()
	_server = host.NewOAuthClient(options)

	mvc.Configure(_server.IrisApp, configureMVC)

	_server.Run()
}

type baseController struct {
	//Ctx     iris.Context
	Session *sessions.Session
	Layout  map[string]interface{}
}

func (c *baseController) BeginRequest(ctx iris.Context) {
	ctx.ViewData("Debug", _server.ConfigProvider.GetBool("OAuthClient.Debug"))

	var isAuth bool
	user := _server.GetUser(ctx)
	if user != nil {
		ctx.ViewData("UserID", user.ID)
		isAuth = true
	}
	ctx.ViewData("IsAuthenticated", isAuth)
	if isAuth {
		ctx.ViewData("Username", user.Username)
		ctx.ViewData("UserRoles", user.Role)
		ctx.ViewData("UserPermissionLevel", user.Level)
		ctx.ViewData("IsAdmin", user.Role == 4)
	}
}
func (c *baseController) EndRequest(ctx iris.Context) {}

type home struct {
	baseController
}

// Get handles GET:/.
func (c *home) Get(ctx iris.Context) mvc.Result {
	return mvc.View{
		Name: "/index.html",
	}
}

func (c *home) GetUsers(ctx iris.Context) {
	httpClient, _ := _server.UserClient(ctx)
	api := newUserAPI(_server.URLProvider)
	buffer := api.GetUsers(httpClient)
	defer api.RecycleBuffer(buffer)
	ctx.Write(buffer.Bytes())
}

func configureMVC(app *mvc.Application) {
	// Root
	rootApp := app.Party("/")
	rootApp.Register(_server.SessionManager.Start)
	rootApp.Handle(new(home))
}

// UserAdminAPI  _
type userAPI struct {
	shttp.APIClient
}

// NewUserAdminAPI _
func newUserAPI(urlProvider surl.IURLProvider) *userAPI {
	r := new(userAPI)
	r.URLProvider = urlProvider
	return r
}

func (x *userAPI) GetUsers(httpClient *http.Client) (r *bytes.Buffer) {
	var err error
	r, err = x.DoBuffer(httpClient, "GET", "https://i.test.com/users", nil, nil)
	if u.LogError(err) {
		return
	}
	return r
}
