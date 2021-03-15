package shttp

import (
	"encoding/json"

	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host/model"
)

func GetUser(ctx IHttpContext, userJsonSessionkey string) (r *model.User) {
	userJson := ctx.GetSessionString(userJsonSessionkey)
	if userJson != "" {
		// 已登录
		err := json.Unmarshal(u.StrToBytes(userJson), &r)
		u.LogError(err)
	}
	return
}
