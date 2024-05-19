package paymentbill

type UploadShoppingInfoParams struct {
	AppId         string
	MerchantId    string
	TradeId       string
	TransactionId string
	PayerUid      string
	PayTotal      int
}

type UploadShoppingInfoRequest struct {
	HeaderContentType string                              `header:"Content-Type"`
	OrderKey          *UploadShoppingInfoRequestOrderKey  `json:"order_key"`
	OrderList         *UploadShoppingInfoRequestOrderList `json:"order_list"`
	Payer             *UploadShoppingInfoRequestPayer     `json:"payer"`
	UploadTime        string                              `json:"upload_time"`
}

type UploadShoppingInfoRequestOrderKey struct {
	OrderNumberType string `json:"order_number_type"`
	TransactionId   string `json:"transaction_id"`
	MchId           string `json:"mchid"`
	OutTradeNo      string `json:"out_trade_no"`
}

type UploadShoppingInfoRequestOrderList struct {
	MerchantOrderNo     string                                                 `json:"merchant_order_no"`
	OrderDetailJumpLink *UploadShoppingInfoRequestOrderListOrderDetailJumpLink `json:"order_detail_jump_link"`
	ItemList            *UploadShoppingInfoRequestOrderListItemList            `json:"item_list"`
}

type UploadShoppingInfoRequestOrderListOrderDetailJumpLink struct {
	AppId string `json:"appid"`
	Path  string `json:"path"`
	Type  string `json:"type"`
}

type UploadShoppingInfoRequestOrderListItemList struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UnitPrice   int    `json:"unit_price"`
	Quantity    int    `json:"quantity"`
}

type UploadShoppingInfoRequestPayer struct {
	OpenId string `json:"openid"`
}

type UploadShoppingInfoResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}
