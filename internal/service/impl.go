package service

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/amazingchow/wechat-payment-callback-service/internal/common/config"
	"github.com/amazingchow/wechat-payment-callback-service/internal/common/logger"
	dao "github.com/amazingchow/wechat-payment-callback-service/internal/extensions/ext_mongo"
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

// 生成平台订单交易ID, 用于标识平台订单.
func (impl *WechatPaymentCallbackServiceImpl) MakeNewPlatformTradeId(
	ctx context.Context, req *proto_gens.MakeNewPlatformTradeIdRequest) (
	resp *proto_gens.MakeNewPlatformTradeIdResponse, err error) {

	_logger := logger.GetGlobalLogger().
		WithField(common.LoggerKeyTraceId, ctx.Value(common.ContextKeyTraceId).(string)).
		WithField(common.LoggerKeySpanId, ctx.Value(common.ContextKeySpanId).(string)).
		WithField(common.LoggerKeyEvent, "MakeNewPlatformTradeId")

	// NOTE: 如果服务需要多实例部署, 请使用分布式ID生成器
	atomic.AddUint64(&(impl.ops), 1)
	tradeId := fmt.Sprintf("%s%02d", time.Now().Format("20060102150405003"), (impl.ops)%100)
	_logger.Debugf("Generated new TradeId: %s.", tradeId)

	resp = &proto_gens.MakeNewPlatformTradeIdResponse{
		TradeId: tradeId,
	}
	return
}

// 创建平台订单和微信预支付订单, 其中平台订单用于标识交易, 微信预支付订单用于发起支付.
// 文档链接: https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter3_5_1.shtml
// NOTE: 如何调用 MakeNewWxPrepayOrder 接口失败, 务必由调用端发起去调用 CloseWxPrepayOrder 接口来关闭微信预支付订单.
func (impl *WechatPaymentCallbackServiceImpl) MakeNewWxPrepayOrder(
	ctx context.Context, req *proto_gens.MakeNewWxPrepayOrderRequest) (
	resp *proto_gens.MakeNewWxPrepayOrderResponse, err error) {

	_logger := logger.GetGlobalLogger().
		WithField(common.LoggerKeyTraceId, ctx.Value(common.ContextKeyTraceId).(string)).
		WithField(common.LoggerKeySpanId, ctx.Value(common.ContextKeySpanId).(string)).
		WithField(common.LoggerKeyEvent, "MakeNewWxPrepayOrder").
		WithField("trade_id", req.TradeId)

	// 参数校验
	if len(req.AppId) == 0 {
		err = status.Error(codes.InvalidArgument, "Empty app_id")
		return
	}
	if _, ok := impl.confSupportedAppIdTable[req.AppId]; !ok {
		err = status.Error(codes.InvalidArgument, "Unsupported app_id")
		return
	}
	if len(req.PayerUid) == 0 {
		err = status.Error(codes.InvalidArgument, "Empty payer_uid")
		return
	}
	if len(req.TradeId) == 0 {
		err = status.Error(codes.InvalidArgument, "Empty trade_id")
		return
	}
	if len(req.ItemDescription) == 0 {
		err = status.Error(codes.InvalidArgument, "Empty item_description")
		return
	}
	if req.ItemAmountTotal <= 0 {
		err = status.Error(codes.InvalidArgument, "Invalid item_amount_total")
		return
	}

	// 1. 生成平台订单, 创建数据库订单记录
	ct := time.Now()
	if innerErr := impl.storage.AddPlatformOrder(ctx, &dao.PlatformOrderModel{
		AppId:           req.AppId,
		MchId:           impl.confMerchantId,
		TradeId:         req.TradeId,
		PayerUid:        req.PayerUid,
		ItemDescription: req.ItemDescription,
		ItemAmountTotal: req.ItemAmountTotal,
		Status:          dao.PaymentStatusCreatePlatformOrder,
		ExpireTime:      ct.Add(time.Duration(config.GetConfig().ServiceInternalConfig.PaymentExpireTimeInMinute) * time.Minute).Unix(),
		CreateTime:      ct.Unix(),
		UpdateTime:      ct.Unix(),
	}); innerErr != nil {
		_logger.WithError(innerErr).Error("Failed to invoke impl.storage.AddPlatformOrder.")
		err = status.Error(codes.Internal, "Failed to create platform-order.")
		return
	}

	// 2. 请求JSAPI下单接口, 创建微信预支付订单, 更新数据库订单记录
	_ = impl.storage.UpdatePlatformOrder(ctx, req.TradeId, dao.PaymentStatusCreateWXPrepayOrder)
	retries := 0
RETRY:
	prepayresp, _, innerErr := impl.svc.Prepay(
		ctx,
		jsapi.PrepayRequest{
			Appid:       core.String(req.AppId),
			Mchid:       core.String(impl.confMerchantId),
			Description: core.String(req.ItemDescription),
			OutTradeNo:  core.String(req.TradeId),
			TimeExpire:  core.Time(ct.Add(time.Duration(config.GetConfig().ServiceInternalConfig.PaymentExpireTimeInMinute) * time.Minute)),
			NotifyUrl:   core.String(impl.confNotifyUrl),
			Amount:      &jsapi.Amount{Total: core.Int64(req.ItemAmountTotal)},
			Payer:       &jsapi.Payer{Openid: core.String(req.PayerUid)},
		},
	)
	if innerErr != nil {
		_logger.WithError(innerErr).Error("Failed to invoke JsapiApiService.Prepay.")

		if core.IsAPIError(innerErr, "SYSTEM_ERROR") ||
			core.IsAPIError(innerErr, "BANK_ERROR") ||
			core.IsAPIError(innerErr, "FREQUENCY_LIMITED") {
			// 处理系统错误/银行系统异常/频率超限, 使用相同参数至多重新调用三次
			if retries < 3 {
				retries += 1
				// NOTE: 重试间隔时间可以根据实际情况调整
				time.Sleep(time.Duration(retries) * time.Second)
				goto RETRY
			} else {
				// 重试次数超限
				err = status.Error(codes.Internal, "WX_PRE_PAY_ERROR")
			}
		} else {
			// 处理其他错误
			err = status.Error(codes.Internal, "WX_PRE_PAY_ERROR")
		}
	}
	if innerErr != nil {
		// 3. 请求JSAPI下单接口失败, 更新数据库订单记录
		_ = impl.storage.UpdatePlatformOrder(ctx, req.TradeId, dao.PaymentStatusCreateWXPrepayOrderFailed)
		return
	}
	// 3. 请求JSAPI下单接口成功, 返回预支付订单标识, 更新数据库订单记录
	_ = impl.storage.UpdatePlatformOrder(ctx, req.TradeId, dao.PaymentStatusGetPrepayId)

	// 4. 生成带签名支付信息
	ts, nonce, pkg := time.Now().Unix(), GenerateNonce(), fmt.Sprintf("prepay_id=%s", *prepayresp.PrepayId)
	signature, innerErr := MakeNewPaymentSignature(impl.confMchPrivateKey, &SignParams{
		AppId:     req.AppId,
		Timestamp: ts,
		Nonce:     nonce,
		Package:   pkg,
	})
	if innerErr != nil {
		// 4. 生成带签名支付信息失败, 更新数据库订单记录
		_ = impl.storage.UpdatePlatformOrder(ctx, req.TradeId, dao.PaymentStatusCreatePaymentSignatureFailed)
		_logger.WithError(innerErr).Error("Failed to invoke MakeNewPaymentSignature.")
		err = status.Error(codes.Internal, "Failed to create payment signature.")
		return
	}
	// 更新数据库订单记录
	_ = impl.storage.UpdatePlatformOrder(ctx, req.TradeId, dao.PaymentStatusCreatePaymentSignature)

	resp = &proto_gens.MakeNewWxPrepayOrderResponse{
		Params: &proto_gens.WxRequestPaymentParams{
			Timestamp: ts,
			Nonce:     nonce,
			Package:   pkg,
			SignType:  SignTypeRSA,
			PaySign:   signature,
		},
	}
	// TODO: 如果N分钟内客户端未通知系统支付成功, 则由系统发起去关闭微信预支付订单

	return
}

// 查询微信支付状态.
// 文档链接: https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter3_5_2.shtml
// NOET: 以下情况需要调用查询接口：
// • 当商户后台、网络、服务器等出现异常，商户系统最终未接收到支付通知（N分钟后）。
// • 调用支付接口后，返回系统错误或未知交易状态情况。
// • 调用付款码支付API后，返回USERPAYING状态。
// • 调用关单或撤销接口API之前，需确认支付状态。
func (impl *WechatPaymentCallbackServiceImpl) QueryWxPaymentStatus(
	ctx context.Context, req *proto_gens.QueryWxPaymentStatusRequest) (
	resp *proto_gens.QueryWxPaymentStatusResponse, err error) {

	_logger := logger.GetGlobalLogger().
		WithField(common.LoggerKeyTraceId, ctx.Value(common.ContextKeyTraceId).(string)).
		WithField(common.LoggerKeySpanId, ctx.Value(common.ContextKeySpanId).(string)).
		WithField(common.LoggerKeyEvent, "QueryWxPaymentStatus").
		WithField("trade_id", req.TradeId)

	retries := 0
RETRY:
	queryorderresp, _, innerErr := impl.svc.QueryOrderByOutTradeNo(
		ctx,
		jsapi.QueryOrderByOutTradeNoRequest{
			OutTradeNo: core.String(req.TradeId),
			Mchid:      core.String(impl.confMerchantId),
		},
	)
	if innerErr != nil {
		_logger.WithError(innerErr).Error("Failed to invoke JsapiApiService.QueryOrderByOutTradeNo.")

		if core.IsAPIError(innerErr, "SYSTEMERROR") ||
			core.IsAPIError(innerErr, "BANKERROR") ||
			core.IsAPIError(innerErr, "FREQUENCY_LIMITED") {
			// 处理系统错误/银行系统异常/频率超限, 使用相同参数至多重新调用三次
			if retries < 3 {
				retries += 1
				// NOTE: 重试间隔时间可以根据实际情况调整
				time.Sleep(time.Duration(retries) * time.Second)
				goto RETRY
			} else {
				// 重试次数超限
				err = status.Error(codes.Internal, "WX_QUERY_ORDER_ERROR")
				return
			}
		} else {
			// 处理其他错误
			err = status.Error(codes.Internal, "WX_QUERY_ORDER_ERROR")
			return
		}
	}

	resp = &proto_gens.QueryWxPaymentStatusResponse{
		AppId:      *(queryorderresp.Appid),
		TradeId:    *(queryorderresp.OutTradeNo),
		TradeState: *(queryorderresp.TradeState),
	}
	if queryorderresp.TransactionId != nil {
		resp.TrxId = *(queryorderresp.TransactionId)
	}
	if queryorderresp.TradeType != nil {
		resp.TradeType = *(queryorderresp.TradeType)
	}
	if queryorderresp.SuccessTime != nil {
		resp.SuccessTime = *(queryorderresp.SuccessTime)
	}

	return
}

// 关闭微信支付订单.
// 文档链接: https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter3_5_3.shtml
// NOET: 以下情况需要调用关单接口：
// • 商户订单支付失败需要生成新单号重新发起支付，要对原订单号调用关单，避免重复支付。
// • 系统下单后，用户支付超时，系统退出不再受理，避免用户继续，请调用关单接口。
// 备注：
// • 关单没有时间限制，建议在订单生成后间隔几分钟（最短5分钟）再调用关单接口，避免出现订单状态同步不及时导致关单失败。
func (impl *WechatPaymentCallbackServiceImpl) CloseWxPrepayOrder(
	ctx context.Context, req *proto_gens.CloseWxPrepayOrderRequest) (
	resp *proto_gens.CloseWxPrepayOrderResponse, err error) {

	_logger := logger.GetGlobalLogger().
		WithField(common.LoggerKeyTraceId, ctx.Value(common.ContextKeyTraceId).(string)).
		WithField(common.LoggerKeySpanId, ctx.Value(common.ContextKeySpanId).(string)).
		WithField(common.LoggerKeyEvent, "CloseWxPrepayOrder").
		WithField("trade_id", req.TradeId)
	_ = _logger

	retries := 0
RETRY:
	_, innerErr := impl.svc.CloseOrder(
		ctx,
		jsapi.CloseOrderRequest{
			OutTradeNo: core.String(req.TradeId),
			Mchid:      core.String(impl.confMerchantId),
		},
	)
	if innerErr != nil {
		_logger.WithError(innerErr).Error("Failed to invoke JsapiApiService.CloseOrder")

		if core.IsAPIError(innerErr, "ORDERNOTEXIST") ||
			core.IsAPIError(innerErr, "ORDER_CLOSED") ||
			core.IsAPIError(innerErr, "MCH_NOT_EXISTS") {
			// 处理订单不存在/订单已关闭/商户号不存在
			resp = &proto_gens.CloseWxPrepayOrderResponse{}
			return
		} else if core.IsAPIError(innerErr, "SYSTEMERROR") ||
			core.IsAPIError(innerErr, "BANKERROR") ||
			core.IsAPIError(innerErr, "FREQUENCY_LIMITED") {
			// 处理系统错误/银行系统异常/频率超限, 使用相同参数至多重新调用三次
			if retries < 3 {
				retries += 1
				// NOTE: 重试间隔时间可以根据实际情况调整
				time.Sleep(time.Duration(retries) * time.Second)
				goto RETRY
			} else {
				// 重试次数超限
				err = status.Error(codes.Internal, "WX_CLOSE_ORDER_ERROR")
				return
			}
		} else {
			// 处理其他错误
			err = status.Error(codes.Internal, "WX_CLOSE_ORDER_ERROR")
			return
		}
	}

	if innerErr != nil {
		// 关闭平台订单失败, 更新数据库订单记录
		_ = impl.storage.UpdatePlatformOrder(ctx, req.TradeId, dao.PaymentStatusClosePlatformOrderFailed)
		return
	}
	// 关闭平台订单成功, 更新数据库订单记录
	_ = impl.storage.UpdatePlatformOrder(ctx, req.TradeId, dao.PaymentStatusClosePlatformOrder)

	resp = &proto_gens.CloseWxPrepayOrderResponse{}
	return
}
