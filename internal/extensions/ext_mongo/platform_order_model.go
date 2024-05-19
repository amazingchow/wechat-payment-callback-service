package extmongo

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PlatformOrderCollection = "platform_orders"
)

type PlatformOrderModel struct {
	Id              primitive.ObjectID `bson:"_id,omitempty"`
	AppId           string             `bson:"app_id"`
	MchId           string             `bson:"merchant_id"`
	TradeId         string             `bson:"trade_id"`
	PayerUid        string             `bson:"payer_uid"`
	ItemDescription string             `bson:"item_description"`
	ItemAmountTotal int64              `bson:"item_amount_total"`
	Status          int                `bson:"status"`
	ExpireTime      int64              `bson:"expire_time"`
	CreateTime      int64              `bson:"create_time"`
	UpdateTime      int64              `bson:"update_time"`
}
