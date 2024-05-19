package extmongo

import "context"

const (
	// 生成平台订单
	PaymentStatusCreatePlatformOrder = iota
	// 生成平台订单失败
	PaymentStatusCreatePlatformOrderFailed
	// 请求JSAPI下单接口, 创建微信预支付订单
	PaymentStatusCreateWXPrepayOrder
	// 请求JSAPI下单接口失败
	PaymentStatusCreateWXPrepayOrderFailed
	// 返回预付单标识
	PaymentStatusGetPrepayId
	// 生成带签名支付信息
	PaymentStatusCreatePaymentSignature
	// 生成带签名支付信息失败
	PaymentStatusCreatePaymentSignatureFailed
	// 关闭平台订单
	PaymentStatusClosePlatformOrder
	// 关闭平台订单失败
	PaymentStatusClosePlatformOrderFailed
	// 异步通知平台支付结果
	PaymentStatusRecvAsyncNotification
	// 保存支付通知
	PaymentStatusStoreAsyncNotification
	// 保存支付通知失败
	PaymentStatusStoreAsyncNotificationFailed
	// 返回告知成功接收处理
	PaymentStatusAckAsyncNotification
)

type PaymentInfoStorage interface {
	AddPlatformOrder(ctx context.Context, order *PlatformOrderModel) (err error)
	UpdatePlatformOrder(ctx context.Context, orderId string, status int) (err error)
	AddPaymentNotification(ctx context.Context, notification *PaymentNotificationModel) (err error)
}
