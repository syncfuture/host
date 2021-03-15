package siris

import (
	"github.com/Lukiya/oauth2go/model"
	"github.com/kataras/iris/v12/view"
	"github.com/syncfuture/go/sconfig"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/host/abstracts"
	"github.com/syncfuture/host/shttp"
)

type (
	IrisBaseServerOptions struct {
		abstracts.BaseHostOptions
		ViewEngine view.Engine
		ViewsDir   string
	}
	OAuthClientOptions struct {
		IrisBaseServerOptions
		AccessDeniedPath    string
		SignInPath          string
		SignInCallbackPath  string
		SignOutPath         string
		SignOutCallbackPath string
		StaticFilesDir      string
		LayoutTemplate      string
		ViewsExtension      string
		SessionName         string
		TokenCookieName     string
		HashKey             string
		BlockKey            string
		OAuth               *abstracts.OAuthOptions
		ContextTokenStore   shttp.IContextTokenStore
		OAuthClientHandler  abstracts.IOAuthClientHandler
	}
	OAuthResourceOptions struct {
		IrisBaseServerOptions
		PublicKeyPath    string
		SigningAlgorithm string
		OAuth            *model.Resource
	}
)

func NewOAuthClientOptions(args ...string) *OAuthClientOptions {
	cp := sconfig.NewJsonConfigProvider(args...)
	var o *OAuthClientOptions
	cp.GetStruct("OAuthClient", &o)
	if o == nil {
		log.Fatal("missing 'OAuthClient' section in configuration")
	}
	o.ConfigProvider = cp
	return o
}
