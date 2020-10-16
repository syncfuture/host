package host

import (
	"crypto/rsa"
	"errors"
	"log"

	"github.com/Lukiya/oauth2go/model"
	jwtgo "github.com/dgrijalva/jwt-go"
	jwtiris "github.com/iris-contrib/middleware/jwt"
	"github.com/kataras/iris/v12"
	"github.com/syncfuture/go/config"
	"github.com/syncfuture/go/rsautil"
	"github.com/syncfuture/go/sslice"
	"github.com/syncfuture/go/u"
)

type (
	OAuthResourceOptions struct {
		IrisBaseServerOptions
		PublicKeyPath    string
		SigningAlgorithm string
		OAuth            *model.Resource
	}

	OAuthResource struct {
		IrisBaseServer
		PublicKey        *rsa.PublicKey
		SigningAlgorithm jwtgo.SigningMethod
		Resource         *model.Resource
	}
)

func NewOAuthResourceOptions(args ...string) *OAuthResourceOptions {
	cp := config.NewJsonConfigProvider(args...)
	var options *OAuthResourceOptions
	cp.GetStruct("OAuthResource", &options)
	if options == nil {
		log.Fatal("missing 'OAuthResource' section in configuration")
	}
	options.ConfigProvider = cp
	return options
}

func NewOAuthResource(options *OAuthResourceOptions) (r *OAuthResource) {
	if options.PublicKeyPath == "" {
		log.Fatal("public key path cannot be empty")
	}
	if options.OAuth == nil {
		log.Fatal("oauth options cannot be nil")
	}
	if options.OAuth.Issuers == nil || len(options.OAuth.Issuers) == 0 {
		log.Fatal("Issuers cannot be empty")
	}
	if options.OAuth.Audiences == nil || len(options.OAuth.Audiences) == 0 {
		log.Fatal("Audiences cannot be empty")
	}
	if options.SigningAlgorithm == "" {
		options.SigningAlgorithm = jwtgo.SigningMethodPS256.Name
	}

	r = new(OAuthResource)
	r.Name = options.Name
	r.configIrisBaseServer(&options.IrisBaseServerOptions)

	for i := range options.OAuth.Issuers {
		options.OAuth.Issuers[i] = r.URLProvider.RenderURL(options.OAuth.Issuers[i])
	}
	for i := range options.OAuth.Audiences {
		options.OAuth.Audiences[i] = r.URLProvider.RenderURL(options.OAuth.Audiences[i])
	}
	r.SigningAlgorithm = jwtgo.GetSigningMethod(options.SigningAlgorithm)
	r.Resource = options.OAuth

	// read public certificate
	cert, err := rsautil.ReadCertFromFile(options.PublicKeyPath)
	u.LogFaltal(err)
	r.PublicKey = cert.PublicKey.(*rsa.PublicKey)

	return
}

func (x *OAuthResource) init(actionGroups ...*[]*Action) {
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
	authMiddleware := &ApiAuthMidleware{
		ActionMap:         x.ActionMap,
		PermissionAuditor: x.PermissionAuditor,
	}

	// 添加中间件
	x.PreMiddlewares = append(x.PreMiddlewares, jwtMiddleware.Serve)
	x.PreMiddlewares = append(x.PreMiddlewares, authMiddleware.Serve)
}

func (x *OAuthResource) validateToken(token *jwtgo.Token) (interface{}, error) {
	claims := token.Claims.(jwtiris.MapClaims)

	// Get iss from JWT and validate against desired iss
	issuer, ok := claims["iss"].(string)
	if !ok {
		return nil, errors.New("issuer is required")
	}
	if !sslice.HasStr(x.Resource.Issuers, issuer) {
		return nil, errors.New("issuer validation failed")
	}

	// Get audience from JWT and validate against desired audience
	var isAudienceValid bool
	if aud, ok := claims["aud"].(string); ok {
		isAudienceValid = sslice.HasStr(x.Resource.Audiences, aud)
	} else if auds, ok := claims["aud"].([]string); ok {
		isAudienceValid = sslice.HasAnyStr(x.Resource.Audiences, auds)
	}

	if !isAudienceValid {
		return nil, errors.New("audience validation failed")
	}

	return x.PublicKey, nil
}

func (x *OAuthResource) Run(actionGroups ...*[]*Action) {
	x.init(actionGroups...)

	x.registerActions()

	if x.ListenAddr == "" {
		log.Fatal("Cannot find 'ListenAddr' config")
	}
	x.IrisApp.Run(iris.Addr(x.ListenAddr))
}
