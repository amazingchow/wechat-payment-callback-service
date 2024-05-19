package paymentbill

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bytedance/sonic/decoder" // https://www.cloudwego.io/zh/docs/hertz/reference/json/

	"github.com/amazingchow/wechat-payment-callback-service/internal/common/logger"
	ext_redis "github.com/amazingchow/wechat-payment-callback-service/internal/extensions/ext_redis"
	"github.com/amazingchow/wechat-payment-callback-service/internal/service/common"
	"github.com/amazingchow/wechat-payment-callback-service/internal/utils/httptools"
)

const CONST_WX_ACCESS_TOKEN_KEY_PREFIX = "miniprogram_backend_service.wx.access_token.appid."

func GetAccessToken(appid string) (token string) {
	if err := ext_redis.GetConnPool().GetBigCache().GetString(
		context.Background(),
		fmt.Sprintf("%s%s", CONST_WX_ACCESS_TOKEN_KEY_PREFIX, appid),
		&token,
	); err != nil {
		logger.GetGlobalLogger().WithError(err).Errorf("Failed to get access_token for app<appid:%s>.", appid)
	}
	return
}

// TODO: 支持错误重试
func UploadShoppingInfo(ctx context.Context, params *UploadShoppingInfoParams) {

	_logger := logger.GetGlobalLogger().
		WithField(common.LoggerKeyTraceId, ctx.Value(common.ContextKeyTraceId).(string)).
		WithField(common.LoggerKeySpanId, ctx.Value(common.ContextKeySpanId).(string)).
		WithField(common.LoggerKeyEvent, "UploadShoppingInfo").
		WithField("trade_id", params.TradeId)

	token := GetAccessToken(params.AppId)
	if len(token) == 0 {
		_logger.Errorf("Failed to invoke UploadShoppingInfo, since access_token for app<app_id:%s> is empty.",
			params.AppId)
		return
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/user-order/orders?access_token=%s", token)
	resp, err := httptools.DoWithCustomClient(
		ctx,
		_HttpCli,
		http.MethodPost,
		url,
		&UploadShoppingInfoRequest{
			HeaderContentType: "application/json",
			OrderKey: &UploadShoppingInfoRequestOrderKey{
				OrderNumberType: "WXPAY_TRADE_NUMBER",
				TransactionId:   params.TransactionId,
				MchId:           params.MerchantId,
				OutTradeNo:      params.TradeId,
			},
			OrderList: &UploadShoppingInfoRequestOrderList{
				MerchantOrderNo: params.TradeId,
				OrderDetailJumpLink: &UploadShoppingInfoRequestOrderListOrderDetailJumpLink{
					AppId: params.AppId,
					Path:  "/index",
					Type:  "MINI_PROGRAM",
				},
				ItemList: &UploadShoppingInfoRequestOrderListItemList{
					Name:      "虚拟商品",
					UnitPrice: params.PayTotal,
					Quantity:  1,
				},
			},
			Payer: &UploadShoppingInfoRequestPayer{
				OpenId: params.PayerUid,
			},
			UploadTime: time.Now().Format(time.RFC3339),
		},
		false,
	)
	if err != nil {
		_logger.WithError(err).Error("Failed to invoke UploadShoppingInfo.")
		return
	}
	defer resp.Body.Close()

	var result UploadShoppingInfoResponse
	decoder.NewStreamDecoder(resp.Body).Decode(&result)
	if result.ErrCode != 0 {
		_logger.Errorf("Failed to invoke UploadShoppingInfo, ErrCode:%d, ErrMsg:%s.",
			result.ErrCode, result.ErrMsg)
	} else {
		_logger.Debug("Invoked UploadShoppingInfo successfully.")
	}
}
