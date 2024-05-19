package service

type AsyncNotificationFromWeChatPay struct {
	/* 通知的唯一ID, 示例值：EV-2018022511223320873 */
	Id string `json:"id"`
	/* 通知创建的时间, 示例值：2015-05-20T13:29:35+08:00 */
	CreateTime string `json:"create_time"`
	/* 通知的事件类型, 支付成功通知为"TRANSACTION.SUCCESS" */
	EventType string `json:"event_type"`
	/* 通知的资源类型, 支付成功通知为"encrypt-resource" */
	ResourceType string `json:"resource_type"`
	/* 通知资源数据 */
	Resource *AsyncNotificationResourceFromWeChatPay `json:"resource"`
	/* 回调摘要, 示例值：支付成功 */
	Summary string `json:"summary"`
}

type AsyncNotificationResourceFromWeChatPay struct {
	/* 原始回调类型, 为"transaction" */
	OriginalType string `json:"original_type"`
	/* 加密算法类型, 目前只支持"AEAD_AES_256_GCM" */
	Algorithm string `json:"algorithm"`
	/* 数据密文 */
	Ciphertext string `json:"ciphertext"`
	/* 附加数据 */
	AssociatedData string `json:"associated_data"`
	/* 随机串 */
	Nonce string `json:"nonce"`
}

type NotificationResource struct {
	/* 应用ID */
	AppId string `json:"appid"`
	/* 商户号 */
	MchId string `json:"mchid"`
	/* 商户订单号 */
	OutTradeNo string `json:"out_trade_no"`
	/* 微信支付订单号 */
	TransactionId string `json:"transaction_id"`
	/* 交易类型, JSAPI: 公众号支付; NATIVE: 扫码支付; APP: APP支付; MICROPAY: 付款码支付; MWEB: H5支付; FACEPAY: 刷脸支付 */
	TradeType string `json:"trade_type"`
	/* 交易状态, SUCCESS: 支付成功; REFUND: 转入退款; NOTPAY: 未支付; CLOSED: 已关闭; REVOKED: 已撤销（付款码支付）; USERPAYING: 用户支付中（付款码支付）; PAYERROR: 支付失败(其他原因，如银行返回失败) */
	TradeState string `json:"trade_state"`
	/* 交易状态描述 */
	TradeStateDesc string `json:"trade_state_desc"`
	/* 付款银行 */
	BankType string `json:"bank_type"`
	/* 支付完成时间 */
	SuccessTime string `json:"success_time"`
	/* 支付者 */
	Payer *NotificationResourcePayer `json:"payer"`
	/* 订单金额 */
	Amount *NotificationResourceAmount `json:"amount"`
}

type NotificationResourcePayer struct {
	/* 用户标识 */
	OpenId string `json:"openid"`
}

type NotificationResourceAmount struct {
	/* 总金额 */
	Total int `json:"total"`
	/* 用户支付金额 */
	PayerTotal int `json:"payer_total"`
	/* 货币类型 */
	Currency string `json:"currency"`
	/* 用户支付币种 */
	PayerCurrency string `json:"payer_currency"`
}

type AckAsyncNotificationFromWeChatPay struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type CommonReponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
