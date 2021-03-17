package siris

import (
	"crypto/rsa"
	"log"

	"github.com/Lukiya/oauth2go/model"
	"github.com/kataras/iris/v12"
	"github.com/pascaldekloe/jwt"
	config "github.com/syncfuture/go/sconfig"
	"github.com/syncfuture/go/srsautil"
	"github.com/syncfuture/go/u"
)

type (
	OAuthResource struct {
		IrisBaseServer
		PublicKey        *rsa.PublicKey
		SigningAlgorithm string
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
	if options.OAuth.ValidIssuers == nil || len(options.OAuth.ValidIssuers) == 0 {
		log.Fatal("Issuers cannot be empty")
	}
	if options.OAuth.ValidAudiences == nil || len(options.OAuth.ValidAudiences) == 0 {
		log.Fatal("Audiences cannot be empty")
	}
	if options.SigningAlgorithm == "" {
		options.SigningAlgorithm = jwt.PS256
	}

	r = new(OAuthResource)
	r.ConfigIrisBaseServer(&options.IrisBaseServerOptions)
	// r.Name = options.Name
	// r.URIKey = options.URIKey
	// r.RouteKey = options.RouteKey
	// r.PermissionKey = options.PermissionKey

	for i := range options.OAuth.ValidIssuers {
		options.OAuth.ValidIssuers[i] = r.URLProvider.RenderURL(options.OAuth.ValidIssuers[i])
	}
	for i := range options.OAuth.ValidAudiences {
		options.OAuth.ValidAudiences[i] = r.URLProvider.RenderURL(options.OAuth.ValidAudiences[i])
	}
	r.SigningAlgorithm = options.SigningAlgorithm
	r.Resource = options.OAuth

	// read public certificate
	cert, err := srsautil.ReadCertFromFile(options.PublicKeyPath)
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

	// jwtMiddleware := jwtiris.New(jwtiris.Config{
	// 	ValidationKeyGetter: x.validateToken,
	// 	SigningMethod:       x.SigningAlgorithm,
	// 	Expiration:          true,
	// })

	// JWT验证中间件
	jwtMiddleware := &JWTMiddleware{
		IssuerSigningKey: x.PublicKey,
		ValidAudiences:   x.Resource.ValidAudiences,
		ValidIssuers:     x.Resource.ValidIssuers,
	}

	// 授权中间件
	authMiddleware := &ApiAuthMidleware{
		ActionMap:         x.ActionMap,
		PermissionAuditor: x.PermissionAuditor,
	}

	// 添加中间件
	x.PreMiddlewares = append(x.PreMiddlewares, jwtMiddleware.Serve)
	x.PreMiddlewares = append(x.PreMiddlewares, authMiddleware.Serve)
}

// func (x *OAuthResource) validateToken(token *jwtgo.Token) (interface{}, error) {
// 	claims := token.Claims.(jwtiris.MapClaims)

// 	// Get iss from JWT and validate against desired iss
// 	issuer, ok := claims["iss"].(string)
// 	if !ok {
// 		return nil, errors.New("issuer is required")
// 	}
// 	if !sslice.HasStr(x.Resource.Issuers, issuer) {
// 		return nil, errors.New("issuer validation failed")
// 	}

// 	// Get audience from JWT and validate against desired audience
// 	var isAudienceValid bool
// 	if aud, ok := claims["aud"].(string); ok {
// 		isAudienceValid = sslice.HasStr(x.Resource.Audiences, aud)
// 	} else if auds, ok := claims["aud"].([]string); ok {
// 		isAudienceValid = sslice.HasAnyStr(x.Resource.Audiences, auds)
// 	}

// 	if !isAudienceValid {
// 		return nil, errors.New("audience validation failed")
// 	}

// 	return x.PublicKey, nil
// }

func (x *OAuthResource) Run(actionGroups ...*[]*Action) {
	x.init(actionGroups...)

	x.RegisterActions()

	if x.ListenAddr == "" {
		log.Fatal("Cannot find 'ListenAddr' config")
	}
	x.IrisApp.Run(iris.Addr(x.ListenAddr))
}
