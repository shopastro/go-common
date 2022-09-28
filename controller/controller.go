package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/shopastro/go-common/common"
	"github.com/shopastro/logs"
	"go.uber.org/zap"
)

type (
	Controller struct {
		localizer *i18n.Localizer
	}

	Response struct {
		Code      int         `json:"code"`
		Status    int         `json:"status"`
		Msg       string      `json:"msg"`
		Data      interface{} `json:"data"`
		Timestamp float64     `json:"timestamp"`
		TraceId   string      `json:"traceId"`
	}
)

const (
	DewuCode           = "DEWU_CODE"
	DewuReleaseVersion = "DEWU_RELEASE_VERSION"
)

func NewController(ctl *Controller) *Controller {
	return ctl
}

func (ctl *Controller) Response(ctx *gin.Context, data interface{}) {
	ctx.Set(DewuCode, http.StatusOK)

	ctx.JSON(http.StatusOK, Response{
		TraceId: common.NewRequest().TraceId(ctx),
		Code:    http.StatusOK,
		Status:  http.StatusOK,
		Msg:     ctl.getLocalize(ctx).i18nLocalize(http.StatusOK),
		Data:    data,
	})
}

func (ctl *Controller) ParamsException(ctx *gin.Context, err error) {
	ctx.Set(DewuCode, 900)

	logs.Logger.Error("[ParamsException]",
		zap.String("uri", ctx.Request.URL.Path),
		zap.Error(err))

	ctx.JSON(http.StatusOK, Response{
		TraceId: common.NewRequest().TraceId(ctx),
		Code:    900,
		Status:  900,
		Msg:     ctl.getLocalize(ctx).i18nLocalize(900),
		Data:    nil,
	})
}

func (ctl *Controller) ServiceException(ctx *gin.Context, err error) {
	ctx.Set(DewuCode, http.StatusInternalServerError)

	logs.Logger.Error("[ServiceException]",
		zap.String("uri", ctx.Request.URL.Path),
		zap.Error(err))

	ctx.JSON(http.StatusOK, Response{
		TraceId: common.NewRequest().TraceId(ctx),
		Code:    http.StatusInternalServerError,
		Status:  http.StatusInternalServerError,
		Msg:     ctl.getLocalize(ctx).i18nLocalize(http.StatusInternalServerError),
		Data:    nil,
	})
}

func (ctl *Controller) ServiceCodeException(ctx *gin.Context, status int, err error, args ...interface{}) {
	ctx.Set(DewuCode, status)

	logs.Logger.Error("[ServiceCodeException]",
		zap.String("uri", ctx.Request.URL.Path),
		zap.Error(err))

	msg := ctl.getLocalize(ctx).i18nLocalize(status)
	resp := Response{
		TraceId: common.NewRequest().TraceId(ctx),
		Code:    status,
		Status:  status,
		Msg:     msg,
		Data:    nil,
	}
	if len(args) > 0 {
		resp.Msg = fmt.Sprintf(msg, args...)
	}
	ctx.JSON(http.StatusOK, resp)
}

func (ctl *Controller) UnauthorizedException(ctx *gin.Context) {
	ctx.Set(DewuCode, http.StatusUnauthorized)

	ctx.JSON(http.StatusUnauthorized, Response{
		TraceId: common.NewRequest().TraceId(ctx),
		Code:    http.StatusUnauthorized,
		Status:  http.StatusUnauthorized,
		Msg:     http.StatusText(http.StatusUnauthorized),
		Data:    nil,
	})
}

func (ctl *Controller) NeedLoginException(ctx *gin.Context, user UserInfoHeader) {

	logs.Logger.Error("[NeedLoginException]",
		zap.String("uri", ctx.Request.URL.Path),
		zap.Any("users", user))

	if user.IsGuest {
		ctx.Set(DewuCode, 7999)
		ctx.JSON(http.StatusUnauthorized, Response{
			TraceId: common.NewRequest().TraceId(ctx),
			Code:    7999,
			Status:  7999,
			Msg:     "游客请登录",
			Data:    nil,
		})
	} else {
		ctx.Set(DewuCode, 700)
		ctx.JSON(http.StatusUnauthorized, Response{
			TraceId: common.NewRequest().TraceId(ctx),
			Code:    700,
			Status:  700,
			Msg:     "请先登录",
			Data:    nil,
		})
	}
}

func (ctl *Controller) Health(ctx *gin.Context) {
	ctx.Set(DewuCode, http.StatusOK)

	ctl.Response(ctx, nil)
}

func (ctl *Controller) i18nLocalize(status int) string {

	return ctl.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: strconv.Itoa(status),
		DefaultMessage: &i18n.Message{
			ID:    strconv.Itoa(status),
			Other: "Internal error in the service",
		},
	})
}

func (ctl *Controller) getLocalize(ctx *gin.Context) *Controller {
	localizer, ok := ctx.Get("Localizer")
	if ok && localizer != nil {
		ctl.localizer = localizer.(*i18n.Localizer)
	}

	return ctl
}
