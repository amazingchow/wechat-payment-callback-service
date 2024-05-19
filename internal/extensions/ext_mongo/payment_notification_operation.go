package extmongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/amazingchow/wechat-payment-callback-service/internal/service/common"
)

func (impl *MongoClientConnPool) AddPaymentNotification(ctx context.Context, notification *PaymentNotificationModel) (
	err error) {

	logger := impl.logger.
		WithField(common.LoggerKeyTraceId, ctx.Value(common.ContextKeyTraceId).(string)).
		WithField(common.LoggerKeySpanId, ctx.Value(common.ContextKeySpanId).(string)).
		WithField(common.LoggerKeyEvent, "AddPaymentNotification")

	// 新建一条来自微信的支付通知
	_, err = impl.collections[PaymentNotificationCollection].InsertOne(
		ctx,
		notification,
		options.InsertOne().
			SetComment("service.wechat_pay_backend_service.storage.mongo.method.add_payment_notification"),
	)
	if err != nil {
		logger.WithError(err).Error(
			"failed to insert one new payment-notification")
		return
	} else {
		logger.Info(
			"insert one new payment-notification")
	}

	return
}
