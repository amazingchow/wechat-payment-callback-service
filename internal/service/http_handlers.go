package service

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	wechatpay_utils "github.com/wechatpay-apiv3/wechatpay-go/utils"

	"github.com/amazingchow/wechat-payment-callback-service/internal/common/logger"
	dao "github.com/amazingchow/wechat-payment-callback-service/internal/extensions/ext_mongo"
	"github.com/amazingchow/wechat-payment-callback-service/internal/service/common"
	payment_bill "github.com/amazingchow/wechat-payment-callback-service/internal/service/payment_bill"
	"github.com/amazingchow/wechat-payment-callback-service/internal/utils/gopool"
)

func (impl *WechatPaymentCallbackServiceImpl) HomeHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, &CommonReponse{
		Code:    0,
		Message: "OK",
	})
}

// 微信支付回调通知
// 文档链接: https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter3_5_5.shtml

// 通知规则：
// • 用户支付完成后，微信会把相关支付结果和用户信息发送给商户，商户需要接收处理该消息，并返回应答。
// • 对后台通知交互时，如果微信收到商户的应答不符合规范或超时，微信认为通知失败，微信会通过一定的策略定期重新发起通知，
//   尽可能提高通知的成功率，但微信不保证通知最终能成功（通知频率为15s/15s/30s/3m/10m/20m/30m/30m/30m/60m/3h/3h/3h/6h/6h - 总计 24h4m）。

// 通知签名：
// • 加密不能保证通知请求来自微信。微信会对发送给商户的通知进行签名，并将签名值放在通知的HTTP头Wechatpay-Signature。
//   商户应当验证签名，以确认请求来自微信，而不是其他的第三方。

// 通知应答：
// • 接收成功：HTTP应答状态码需返回200或204，无需返回应答报文。
// • 接收失败：HTTP应答状态码需返回5XX或4XX，同时需返回应答报文。

// 注意：
// • 同样的通知可能会多次发送给商户系统。商户系统必须能够正确处理重复的通知。
//   推荐的做法是，当商户系统收到通知进行处理时，先检查对应业务数据的状态，并判断该通知是否已经处理。
//   如果未处理，则再进行处理；如果已处理，则直接返回结果成功。
//   在对业务数据进行状态检查和处理之前，要采用数据锁进行并发控制，以避免函数重入造成的数据混乱。
// • 如果在所有通知频率后没有收到微信侧回调，商户应调用查询订单接口确认订单状态。

// 特别提醒：
// • 商户系统对于开启结果通知的内容一定要做签名验证，并校验通知的信息是否与商户侧的信息一致，防止数据泄露导致出现“假通知”，造成资金损失。
func (impl *WechatPaymentCallbackServiceImpl) NotifyHandler(ctx *gin.Context) {

	_logger := logger.GetGlobalLogger().
		WithField(common.LoggerKeyTraceId, ctx.Value(common.ContextKeyTraceId).(string)).
		WithField(common.LoggerKeySpanId, ctx.Value(common.ContextKeySpanId).(string)).
		WithField(common.LoggerKeyEvent, "NotifyHandler")

	// 1. 解析异步通知
	var notification AsyncNotificationFromWeChatPay
	if err := ctx.BindJSON(&notification); err != nil {
		_logger.WithError(err).Error("Failed to parse AsyncNotificationFromWeChatPay.")
		ctx.JSON(http.StatusBadRequest, &AckAsyncNotificationFromWeChatPay{
			Code:    "FAIL",
			Message: "通知解析失败",
		})
		return
	}
	if notification.Resource == nil {
		_logger.Error("Empty AsyncNotificationFromWeChatPay.Resource.")
		ctx.JSON(http.StatusBadRequest, &AckAsyncNotificationFromWeChatPay{
			Code:    "FAIL",
			Message: "通知资源为空",
		})
		return
	}
	plaintext, err := wechatpay_utils.DecryptAES256GCM(
		impl.confMchAPIv3Key,
		notification.Resource.AssociatedData,
		notification.Resource.Nonce,
		notification.Resource.Ciphertext,
	)
	if err != nil {
		_logger.WithError(err).Error("Failed to decode AsyncNotificationFromWeChatPay.Resource.")
		ctx.JSON(http.StatusBadRequest, &AckAsyncNotificationFromWeChatPay{
			Code:    "FAIL",
			Message: "通知资源解密失败",
		})
		return
	}
	var notificationResource NotificationResource
	if err := json.Unmarshal([]byte(plaintext), &notificationResource); err != nil {
		_logger.WithError(err).Error("Failed to parse NotificationResource.")
		ctx.JSON(http.StatusBadRequest, &AckAsyncNotificationFromWeChatPay{
			Code:    "FAIL",
			Message: "通知资源解析失败",
		})
		return
	}

	// TODO: 2. 使用HTTP头Wechatpay-Signature来验证签名，以确认请求来自微信，而不是其他的第三方
	signHeader := ctx.Request.Header["Wechatpay-Signature"]
	_logger.Debugf("### Wechatpay-Signature: %s", signHeader)

	// 3. 异步通知平台支付结果, 更新数据库订单记录
	_ = impl.storage.UpdatePlatformOrder(ctx, notificationResource.OutTradeNo, dao.PaymentStatusRecvAsyncNotification)

	// 4. 持久化支付通知, 用于离线对账
	if innerErr := impl.storage.AddPaymentNotification(ctx, &dao.PaymentNotificationModel{
		NotifyId:                    notification.Id,
		CreateTime:                  notification.CreateTime,
		EventType:                   notification.EventType,
		ResourceAppId:               notificationResource.AppId,
		ResourceMchId:               notificationResource.MchId,
		TradeId:                     notificationResource.OutTradeNo,
		ResourceTransactionId:       notificationResource.TransactionId,
		ResourceTradeType:           notificationResource.TradeType,
		ResourceTradeState:          notificationResource.TradeState,
		ResourceTradeStateDesc:      notificationResource.TradeStateDesc,
		ResourceBankType:            notificationResource.BankType,
		ResourceSuccessTime:         notificationResource.SuccessTime,
		ResourcePayerOpenId:         notificationResource.Payer.OpenId,
		ResourceAmountTotal:         notificationResource.Amount.Total,
		ResourceAmountPayerTotal:    notificationResource.Amount.PayerTotal,
		ResourceAmountCurrency:      notificationResource.Amount.Currency,
		ResourceAmountPayerCurrency: notificationResource.Amount.PayerCurrency,
		Summary:                     notification.Summary,
	}); innerErr != nil {
		_logger.WithError(innerErr).Error("Failed to invoke impl.storage.AddPaymentNotification.")
		// 保存支付通知失败, 更新数据库订单记录
		_ = impl.storage.UpdatePlatformOrder(ctx, notificationResource.OutTradeNo, dao.PaymentStatusStoreAsyncNotificationFailed)
	} else {
		// 保存支付通知成功, 更新数据库订单记录
		_ = impl.storage.UpdatePlatformOrder(ctx, notificationResource.OutTradeNo, dao.PaymentStatusStoreAsyncNotification)
	}

	// NOTE: 支持在接收到支付通知后上传订单信息
	// 用户后续可以从微信「我」-「服务」-「钱包」-「账单」中进入，也可以从支付凭证消息进入账单详情页回溯已购物的订单。
	gopool.Go(func() {
		_ctx := context.WithValue(
			context.WithValue(
				context.Background(),
				common.ContextKeyTraceId,
				ctx.Value(common.ContextKeyTraceId).(string),
			),
			common.ContextKeySpanId,
			ctx.Value(common.ContextKeySpanId).(string),
		)
		payment_bill.UploadShoppingInfo(_ctx, &payment_bill.UploadShoppingInfoParams{
			AppId:         notificationResource.AppId,
			MerchantId:    notificationResource.MchId,
			TradeId:       notificationResource.OutTradeNo,
			TransactionId: notificationResource.TransactionId,
			PayerUid:      notificationResource.Payer.OpenId,
			PayTotal:      notificationResource.Amount.PayerTotal,
		})
	})

	// 5. 返回告知成功接收处理, 更新数据库订单记录
	_ = impl.storage.UpdatePlatformOrder(ctx, notificationResource.OutTradeNo, dao.PaymentStatusAckAsyncNotification)
	ctx.JSON(http.StatusOK, &AckAsyncNotificationFromWeChatPay{
		Code:    "SUCCESS",
		Message: "",
	})
}
