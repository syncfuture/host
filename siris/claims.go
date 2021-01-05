package siris

import (
	"strconv"

	"github.com/kataras/iris/v12/context"
	"github.com/pascaldekloe/jwt"
)

func GetClaimInt64(claimName string, ctx context.Context) int64 {
	str := GetClaimString(claimName, ctx)
	r, _ := strconv.ParseInt(str, 10, 64)
	return r
}

func GetClaimString(claimName string, ctx context.Context) string {
	j := ctx.Values().Get("jwt")
	if j != nil {
		if token, ok := j.(*jwt.Claims); ok {
			if str, ok := token.Set[claimName].(string); ok && str != "" {
				return str
			}
		}
	}

	return ""
}
