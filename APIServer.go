package host

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"log"

	"github.com/Lukiya/oauth2go/model"
	jwtgo "github.com/dgrijalva/jwt-go"
	jwtiris "github.com/iris-contrib/middleware/jwt"
	"github.com/kataras/iris/v12"
	"github.com/syncfuture/go/config"
	"github.com/syncfuture/go/rsautil"
	"github.com/syncfuture/go/u"
)

type (
	APIServerOptions struct {
		IrisBaseServerOptions
		PublicKeyPath    string
		SigningAlgorithm string
		OAuth            *model.Resource
	}

	APIServer struct {
		IrisBaseServer
		PublicKey        *rsa.PublicKey
		SigningAlgorithm jwtgo.SigningMethod
		Resource         *model.Resource
	}
)

func NewAPIServerOptions(args ...string) *APIServerOptions {
	cp := config.NewJsonConfigProvider(args...)
	var options *APIServerOptions
	cp.GetStruct("APIServer", &options)
	if options == nil {
		log.Fatal("missing 'APIServer' section in configuration")
	}
	options.ConfigProvider = cp
	return options
}

func NewAPIServer(options *APIServerOptions) (r *APIServer) {
	if options.PublicKeyPath == "" {
		log.Fatal("public key path cannot be empty")
	}
	if options.OAuth == nil {
		log.Fatal("oauth options cannot be nil")
	}
	if options.OAuth.Issuers == nil || len(options.OAuth.Issuers) == 0 {
		log.Fatal("issuers cannot be empty")
	}
	if options.OAuth.Scopes == nil || len(options.OAuth.Scopes) == 0 {
		log.Fatal("scopes cannot be empty")
	}
	if options.SigningAlgorithm == "" {
		options.SigningAlgorithm = jwtgo.SigningMethodPS256.Name
	}

	r = new(APIServer)
	r.Name = options.Name
	r.configIrisBaseServer(&options.IrisBaseServerOptions)

	r.SigningAlgorithm = jwtgo.GetSigningMethod(options.SigningAlgorithm)
	r.Resource = options.OAuth

	// read public certificate
	cert, err := rsautil.ReadPrivateKeyFromFile(options.PublicKeyPath)
	u.LogFaltal(err)
	r.PublicKey = &cert.PublicKey

	return
}

func (x *APIServer) init(actionGroups ...*[]*Action) {
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
	if !x.Resource.HasIssuer(claims["iss"].(string)) {
		return nil, fmt.Errorf("cannot validate iss claim")
	}

	// Get audience from JWT and validate against desired audience
	var isAudienceValid bool
	if aud, ok := claims["aud"].(string); ok {
		isAudienceValid = x.Resource.HasScope(aud)
	} else if auds, ok := claims["aud"].([]string); ok {
		isAudienceValid = x.Resource.HasAnyScopes(auds)
	}

	if !isAudienceValid {
		return nil, errors.New("cannot validate audience claim")
	}

	return x.PublicKey, nil
}

func (x *APIServer) Run(actionGroups ...*[]*Action) {
	x.init(actionGroups...)

	x.registerActions()

	if x.ListenAddr == "" {
		log.Fatal("Cannot find 'ListenAddr' config")
	}
	x.WebServer.Run(iris.Addr(x.ListenAddr))
}
