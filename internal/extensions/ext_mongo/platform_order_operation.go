package extmongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/amazingchow/wechat-payment-callback-service/internal/service/common"
)

func (impl *MongoClientConnPool) AddPlatformOrder(ctx context.Context, order *PlatformOrderModel) (
	err error) {

	logger := impl.logger.
		WithField(common.LoggerKeyTraceId, ctx.Value(common.ContextKeyTraceId).(string)).
		WithField(common.LoggerKeySpanId, ctx.Value(common.ContextKeySpanId).(string)).
		WithField(common.LoggerKeyEvent, "AddPlatformOrder")

	// 新建一条平台订单
	_, err = impl.collections[PlatformOrderCollection].InsertOne(
		ctx,
		order,
		options.InsertOne().
			SetComment("service.wechat_pay_backend_service.storage.mongo.method.add_platform_order"),
	)
	if err != nil {
		logger.WithError(err).Error(
			"failed to insert one new platform-order")
		return
	} else {
		logger.Info(
			"insert one new platform-order")
	}

	return
}

func (impl *MongoClientConnPool) UpdatePlatformOrder(ctx context.Context, tradeId string, status int) (
	err error) {

	logger := impl.logger.
		WithField(common.LoggerKeyTraceId, ctx.Value(common.ContextKeyTraceId).(string)).
		WithField(common.LoggerKeySpanId, ctx.Value(common.ContextKeySpanId).(string)).
		WithField(common.LoggerKeyEvent, "UpdatePlatformOrder")

	// 更新平台订单状态
	if _, err = impl.collections[PlatformOrderCollection].UpdateOne(
		ctx,
		bson.M{"trade_id": tradeId},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "status", Value: status},
			{Key: "update_time", Value: time.Now().Unix()},
		}}},
		options.Update().
			SetComment("service.wechat_pay_backend_service.storage.mongo.method.update_platform_order"),
	); err != nil {
		logger.WithError(err).Errorf(
			"failed to update platform-order(trade-id:%s)",
			tradeId,
		)
		return
	} else {
		logger.Infof(
			"update platform-order(trade-id:%s)",
			tradeId,
		)
	}

	return
}
