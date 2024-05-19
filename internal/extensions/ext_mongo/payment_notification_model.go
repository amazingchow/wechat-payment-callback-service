package extmongo

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PaymentNotificationCollection = "payment_notifications"
)

type PaymentNotificationModel struct {
	Id                          primitive.ObjectID `bson:"_id,omitempty"`
	NotifyId                    string             `bson:"notify_id"`
	CreateTime                  string             `bson:"create_time"`
	EventType                   string             `bson:"event_type"`
	ResourceAppId               string             `bson:"resource_appid"`
	ResourceMchId               string             `bson:"resource_mchid"`
	TradeId                     string             `bson:"trade_id"`
	ResourceTransactionId       string             `bson:"resource_transaction_id"`
	ResourceTradeType           string             `bson:"resource_trade_type"`
	ResourceTradeState          string             `bson:"resource_trade_state"`
	ResourceTradeStateDesc      string             `bson:"resource_trade_state_desc"`
	ResourceBankType            string             `bson:"resource_bank_type"`
	ResourceSuccessTime         string             `bson:"resource_success_time"`
	ResourcePayerOpenId         string             `bson:"resource_payer_openid"`
	ResourceAmountTotal         int                `bson:"resource_amount_total"`
	ResourceAmountPayerTotal    int                `bson:"resource_amount_payer_total"`
	ResourceAmountCurrency      string             `bson:"resource_amount_currency"`
	ResourceAmountPayerCurrency string             `bson:"resource_amount_payer_currency"`
	Summary                     string             `bson:"summary"`
}
