package service

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"

	wechatpay_utils "github.com/wechatpay-apiv3/wechatpay-go/utils"
)

const (
	NonceSymbols = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	NonceLength  = 24
)

func GenerateNonce() string {
	bytes := make([]byte, NonceLength)
	rand.Read(bytes)
	l := byte(len(NonceSymbols))
	for i, b := range bytes {
		bytes[i] = NonceSymbols[b%l]
	}
	return string(bytes)
}

type SignParams struct {
	AppId     string
	Timestamp int64
	Nonce     string
	Package   string
}

const (
	SignTypeRSA = "RSA"
)

func MakeNewPaymentSignature(privKey *rsa.PrivateKey, signParams *SignParams) (signature string, err error) {
	source := fmt.Sprintf("%s\n%d\n%s\n%s\n",
		signParams.AppId,
		signParams.Timestamp,
		signParams.Nonce,
		signParams.Package,
	)
	return wechatpay_utils.SignSHA256WithRSA(source, privKey)
}
