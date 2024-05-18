package service

var impl *WechatPaymentCallbackServiceImpl

type WechatPaymentCallbackServiceImpl struct {}

func SetupWechatPaymentCallbackServiceImpl() {
	impl = &WechatPaymentCallbackServiceImpl{}
}

func GetWechatPaymentCallbackServiceImpl() *WechatPaymentCallbackServiceImpl {
	return impl
}

func CloseWechatPaymentCallbackServiceImpl() {}
