package service

import (
	"context"
	"crypto/rsa"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	wechatpay_utils "github.com/wechatpay-apiv3/wechatpay-go/utils"

	"github.com/amazingchow/wechat-payment-callback-service/internal/common/config"
	"github.com/amazingchow/wechat-payment-callback-service/internal/common/logger"
	dao "github.com/amazingchow/wechat-payment-callback-service/internal/extensions/ext_mongo"
)

var impl *WechatPaymentCallbackServiceImpl

type WechatPaymentCallbackServiceImpl struct {
	ops uint64

	confMchPrivateKey       *rsa.PrivateKey
	confSupportedAppIdTable map[string]struct{}
	confMerchantId          string
	confMchAPIv3Key         string
	confNotifyUrl           string

	svc     *jsapi.JsapiApiService
	storage dao.PaymentInfoStorage
}

func SetupWechatPaymentCallbackServiceImpl() {
	impl = &WechatPaymentCallbackServiceImpl{}

	var (
		mchID           string = config.GetConfig().ServiceInternalConfig.MerchantID           // 商户号
		mchCertSerialNo string = config.GetConfig().ServiceInternalConfig.MerchantCertSerialNo // 商户证书序列号
		mchAPIv3Key     string = config.GetConfig().ServiceInternalConfig.MerchantAPIv3Key     // 商户APIv3密钥
	)

	// 1. 加载商户API私钥, 商户API私钥用于生成请求的签名
	mchPrivateKey, err := wechatpay_utils.LoadPrivateKeyWithPath(config.GetConfig().ServiceInternalConfig.MerchantAPIv3SecretCertPath)
	if err != nil {
		logger.GetGlobalLogger().WithError(err).Fatalf("Failed to load merchant private key.")
	}
	impl.confMchPrivateKey = mchPrivateKey
	// 2. 使用商户API私钥等信息来初始化客户端, 并使它具有自动定时获取微信支付平台证书的能力
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(mchID, mchCertSerialNo, mchPrivateKey, mchAPIv3Key),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		logger.GetGlobalLogger().WithError(err).Fatal("Failed to create wechat pay client.")
	}
	// 3. 初始化支付服务
	impl.svc = &jsapi.JsapiApiService{Client: client}

	// 4. 初始化配置
	supportedAppList := config.GetConfig().ServiceInternalConfig.SupportedAppList
	impl.confSupportedAppIdTable = make(map[string]struct{}, len(supportedAppList))
	for i := 0; i < len(supportedAppList); i++ {
		impl.confSupportedAppIdTable[supportedAppList[i].AppID] = struct{}{}
	}
	impl.confMerchantId = mchID
	impl.confMchAPIv3Key = mchAPIv3Key
	impl.confNotifyUrl = config.GetConfig().ServiceInternalConfig.PaymentCallbackNotifyURL

	// 5. 初始化数据库连接
	dao.InitConnPool(&(config.GetConfig().ServiceInternalConfig.Storage))
	impl.storage = dao.GetConnPool()
}

func GetWechatPaymentCallbackServiceImpl() *WechatPaymentCallbackServiceImpl {
	return impl
}

func CloseWechatPaymentCallbackServiceImpl() {
	dao.CloseConnPool()
}
