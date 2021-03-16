package abstracts

import (
	"github.com/Lukiya/oauth2go"
	log "github.com/syncfuture/go/slog"
	"golang.org/x/oauth2"
)

type (
	OAuthOptions struct {
		*oauth2.Config
		PkceRequired       bool
		EndSessionEndpoint string
		SignOutRedirectURL string
		ClientCredential   *oauth2go.ClientCredential
	}
)

func (x *OAuthOptions) BuildOAuthOptions() {
	if x.Endpoint.AuthURL == "" {
		log.Fatal("OAuth.Endpoint.AuthURL cannot be empty")
	}
	if x.Endpoint.TokenURL == "" {
		log.Fatal("OAuth.Endpoint.TokenURL cannot be empty")
	}
	if x.RedirectURL == "" {
		log.Fatal("OAuth.RedirectURL cannot be empty")
	}
	if x.SignOutRedirectURL == "" {
		log.Fatal("OAuth.SignOutRedirectURL cannot be empty")
	}
	if x.EndSessionEndpoint == "" {
		log.Fatal("OAuth.EndSessionEndpoint cannot be empty")
	}
}
