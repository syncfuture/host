package abstracts

import (
	"github.com/Lukiya/oauth2go"
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/surl"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
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

func (x *OAuthOptions) buildOAuthOptions(urlProvider surl.IURLProvider) {
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

	if urlProvider != nil {
		x.Endpoint.AuthURL = urlProvider.RenderURL(x.Endpoint.AuthURL)
		x.Endpoint.TokenURL = urlProvider.RenderURL(x.Endpoint.TokenURL)
		x.EndSessionEndpoint = urlProvider.RenderURL(x.EndSessionEndpoint)
		x.RedirectURL = urlProvider.RenderURL(x.RedirectURL)
		x.SignOutRedirectURL = urlProvider.RenderURL(x.SignOutRedirectURL)
	}
	// if x.ClientID == "" {
	// 	log.Fatal("OAuth.ClientID cannot be empty")
	// }
	// if x.ClientSecret == "" {
	// 	log.Fatal("OAuth.ClientSecret cannot be empty")
	// }

	x.ClientCredential = &oauth2go.ClientCredential{
		Config: &clientcredentials.Config{
			ClientID:     x.ClientID,
			ClientSecret: x.ClientSecret,
			TokenURL:     x.Endpoint.TokenURL,
			Scopes:       x.Scopes,
		},
	}
}
