package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shopastro/logs"
	"go.uber.org/zap"
)

type (
	UserInfoHeader struct {
		Uuid    string `json:"uuid"`
		UserId  uint64 `json:"userId"`
		IsGuest bool   `json:"isGuest"`
	}
)

const (
	TokenName          = "X-Auth-Token"
	StandardClaimsName = "standardClaims"
	BearerTokenPrefix  = "Bearer "
	PoizonUuid         = "POIZON-UUID"
	PoizonUserId       = "POIZON-USERID"
	PoizonIsGuest      = "POIZON-ISGUEST"
)

func (ctl *Controller) GetTokenInfo(ctx *gin.Context) UserInfoHeader {
	userId, err := strconv.ParseInt(ctx.GetHeader(PoizonUserId), 10, 64)
	if err != nil {
		logs.Logger.Error("[GetTokenInfo]",
			zap.String("uri", ctx.Request.RequestURI),
			zap.Error(err))
	}

	IsGuest, err := strconv.ParseBool(ctx.GetHeader(PoizonIsGuest))
	if err != nil {
		IsGuest = true
		logs.Logger.Error("[GetTokenInfo IsGuest]",
			zap.String("uri", ctx.Request.RequestURI),
			zap.Error(err))
	}

	return UserInfoHeader{
		Uuid:    ctx.GetHeader(PoizonUuid),
		UserId:  uint64(userId),
		IsGuest: IsGuest,
	}
}
