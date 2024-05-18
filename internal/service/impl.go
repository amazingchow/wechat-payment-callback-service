package service

import (
	"context"

	"github.com/amazingchow/wechat-payment-callback-service/internal/common/logger"
	"github.com/amazingchow/wechat-payment-callback-service/internal/proto_gens"
	"github.com/amazingchow/wechat-payment-callback-service/internal/service/common"
)

func (impl *WechatPaymentCallbackServiceImpl) Ping(
	ctx context.Context, req *proto_gens.PingRequest) (
	resp *proto_gens.PongResponse, err error) {

	_logger := logger.GetGlobalLogger().
		WithField(common.LoggerKeyTraceId, ctx.Value(common.ContextKeyTraceId).(string)).
		WithField(common.LoggerKeySpanId, ctx.Value(common.ContextKeySpanId).(string)).
		WithField(common.LoggerKeyEvent, "Ping")
	_ = _logger

	resp = &proto_gens.PongResponse{}
	return
}
