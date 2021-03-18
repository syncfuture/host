package sjwt

import (
	"strconv"

	"github.com/pascaldekloe/jwt"
	"github.com/syncfuture/host/abstracts"
)

func GetClaims(ctx abstracts.IHttpContext) *jwt.Claims {
	j, ok := ctx.GetItem(abstracts.Item_JWT).(*jwt.Claims)
	if ok {
		return j
	}
	return nil
}

func GetClaimValue(ctx abstracts.IHttpContext, claimName string) interface{} {
	j := GetClaims(ctx)
	if j != nil {
		if v, ok := j.Set[claimName]; ok {
			return v
		}
	}
	return nil
}

func GetClaimString(ctx abstracts.IHttpContext, claimName string) string {
	v, ok := GetClaimValue(ctx, claimName).(string)
	if ok {
		return v
	}
	return v
}

func GetClaimInt64(ctx abstracts.IHttpContext, claimName string) int64 {
	v := GetClaimValue(ctx, claimName)
	switch value := v.(type) {
	case int64:
		return value
	case string:
		r, _ := strconv.ParseInt(value, 10, 64)
		return r
	default:
		return 0
	}
}
