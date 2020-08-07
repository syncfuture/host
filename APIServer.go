package host

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	jwtgo "github.com/dgrijalva/jwt-go"
	jwtiris "github.com/iris-contrib/middleware/jwt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/syncfuture/go/config"
	"github.com/syncfuture/go/rsautil"
	"github.com/syncfuture/go/u"
)

type OAuth2Options struct {
	Issuer string
	Scopes []string
}

func (x *OAuth2Options) HasAudience(audience string) bool {
	for _, aud := range x.Scopes {
		if audience == aud {
			return true
		}
	}
	return false
}

type APIServerOptions struct {
	Name             string
	PublicKeyPath    string
	ListenAddr       string
	OAuth            *OAuth2Options
	SigningAlgorithm jwtgo.SigningMethod
}

func NewAPIServer(configProvider config.IConfigProvider, options *APIServerOptions) (r *APIServer) {
	if options.PublicKeyPath == "" {
		log.Fatal("public key path cannot be empty")
	}
	if options.OAuth == nil {
		log.Fatal("oauth options cannot be nil")
	}
	if options.OAuth.Issuer == "" {
		log.Fatal("issuer cannot be empty")
	}
	if options.SigningAlgorithm == nil {
		options.SigningAlgorithm = jwtgo.SigningMethodPS256.SigningMethodRSA
	}

	// read public certificate
	// cert, err := rsautil.ReadCertFromFile(options.PublicKeyPath)
	cert, err := rsautil.ReadPrivateKeyFromFile(options.PublicKeyPath)
	u.LogFaltal(err)
	publicKey := cert.PublicKey

	// create pointer
	r = new(APIServer)
	r.Name = options.Name
	r.PublicKey = &publicKey
	r.SigningAlgorithm = options.SigningAlgorithm
	r.OAuthOptions = options.OAuth
	r.ListenAddr = options.ListenAddr

	// common setup
	r.setup(configProvider)
	if r.URLProvider == nil {
		log.Fatal("URIProvider cannot be nil")
	}
	if r.RoutePermissionProvider == nil {
		log.Fatal("RoutePermissionProvider cannot be nil")
	}
	if r.PermissionAuditor == nil {
		log.Fatal("PermissionAuditor cannot be nil")
	}

	return
}

type APIServer struct {
	ServerBase
	PublicKey        *rsa.PublicKey
	PreMiddlewares   []context.Handler
	ActionMap        *map[string]*Action
	SigningAlgorithm jwtgo.SigningMethod
	OAuthOptions     *OAuth2Options
}

func (x *APIServer) Init(actionGroups ...*[]*Action) {
	actionMap := make(map[string]*Action)

	for _, actionGroup := range actionGroups {
		for _, action := range *actionGroup {
			actionMap[action.Route] = action
		}
	}
	x.ActionMap = &actionMap

	// JWT验证中间件
	jwtMiddleware := jwtiris.New(jwtiris.Config{
		ValidationKeyGetter: x.validateToken,
		SigningMethod:       x.SigningAlgorithm,
		Expiration:          true,
	})

	// 授权中间件
	authMiddleware := &AuthMidleware{
		ActionMap:         x.ActionMap,
		PermissionAuditor: x.PermissionAuditor,
	}

	// 添加中间件
	x.PreMiddlewares = append(x.PreMiddlewares, jwtMiddleware.Serve)
	x.PreMiddlewares = append(x.PreMiddlewares, authMiddleware.Serve)
}

func (x *APIServer) validateToken(token *jwtgo.Token) (interface{}, error) {
	claims := token.Claims.(jwtiris.MapClaims)

	// Get iss from JWT and validate against desired iss
	if claims["iss"].(string) != x.OAuthOptions.Issuer {
		return nil, fmt.Errorf("cannot validate iss claim")
	}

	// Get audience from JWT and validate against desired audience
	var isAudienceValid bool
	if aud, _ := claims["aud"].(string); x.OAuthOptions.HasAudience(aud) {
		isAudienceValid = true
	} else if auds, ok := claims["aud"].([]string); ok {
		for _, aud := range auds {
			if x.OAuthOptions.HasAudience(aud) {
				isAudienceValid = true
				break
			}
		}
	}

	if !isAudienceValid {
		return nil, errors.New("cannot validate audience claim")
	}

	return x.PublicKey, nil
}

func (x *APIServer) Run() {
	x.registerActions()

	if x.ListenAddr == "" {
		log.Fatal("Cannot find 'ListenAddr' config")
	}
	x.WebServer.Run(iris.Addr(x.ListenAddr))
}

func (x *APIServer) registerActions() {
	for name, action := range *x.ActionMap {
		handlers := append(x.PreMiddlewares, action.Handler)
		x.registerAction(name, handlers...)
	}
}

func (x *APIServer) registerAction(name string, handlers ...context.Handler) {
	index := strings.Index(name, "/")
	method := name[:index]
	path := name[index:]

	switch method {
	case http.MethodPost:
		x.WebServer.Post(path, handlers...)
		break
	case http.MethodGet:
		x.WebServer.Get(path, handlers...)
		break
	case http.MethodPut:
		x.WebServer.Put(path, handlers...)
		break
	case http.MethodPatch:
		x.WebServer.Patch(path, handlers...)
		break
	case http.MethodDelete:
		x.WebServer.Delete(path, handlers...)
		break
	default:
		panic("does not support method " + method)
	}
}
