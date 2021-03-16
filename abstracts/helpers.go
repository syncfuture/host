package abstracts

import (
	"net/http"

	oauth2core "github.com/Lukiya/oauth2go/core"
	"github.com/syncfuture/go/srand"
	"github.com/syncfuture/host/shttp"
	"golang.org/x/oauth2"
)

// func getRoutes(handlerName string) (string, string, string) {
// 	array := strings.Split(handlerName, ".")
// 	return array[0], array[1], array[2]
// }

func redirectAuthorizeEndpoint(ctx shttp.IHttpContext, oauthOptions *OAuthOptions, returnURL string) {
	state := srand.String(32)
	ctx.SetSession(state, returnURL)
	if oauthOptions.PkceRequired {
		codeVerifier := oauth2core.Random64String()
		codeChanllenge := oauth2core.ToSHA256Base64URL(codeVerifier)
		ctx.SetSession(oauth2core.Form_CodeVerifier, codeVerifier)
		ctx.SetSession(oauth2core.Form_CodeChallengeMethod, oauth2core.Pkce_S256)
		codeChanllengeParam := oauth2.SetAuthURLParam(oauth2core.Form_CodeChallenge, codeChanllenge)
		codeChanllengeMethodParam := oauth2.SetAuthURLParam(oauth2core.Form_CodeChallengeMethod, oauth2core.Pkce_S256)
		ctx.Redirect(oauthOptions.AuthCodeURL(state, codeChanllengeParam, codeChanllengeMethodParam), http.StatusFound)
	} else {
		ctx.Redirect(oauthOptions.AuthCodeURL(state), http.StatusFound)
	}
}
