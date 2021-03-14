package abstracts

import (
	"log"

	"github.com/Lukiya/oauth2go"
	"github.com/kataras/iris/v12/view"
	"github.com/syncfuture/go/sconfig"
	"golang.org/x/oauth2"
)

type (
	OAuthOptions struct {
		oauth2.Config
		PkceRequired       bool
		EndSessionEndpoint string
		SignOutRedirectURL string
		ClientCredential   *oauth2go.ClientCredential
	}
	IrisBaseServerOptions struct {
		BaseServerOptions
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
		OAuth               *OAuthOptions
		OAuthClientHandler  IOAuthClientHandler
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
