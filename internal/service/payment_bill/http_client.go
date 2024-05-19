package paymentbill

import (
	"net/http"
	"time"
)

var _HttpCli *http.Client

func init() {
	_RoundTripper := http.DefaultTransport
	_HttpTransportPtr, ok := _RoundTripper.(*http.Transport)
	if !ok {
		panic("_RoundTripper is not an *http.Transport")
	}
	_HttpTransport := *_HttpTransportPtr
	// In case of "Thousands of connections in the TIME_WAIT state,
	// eventually, service will run out of ephemeral ports and
	// not be able to open new client connections."
	_HttpTransport.MaxIdleConns = 32
	_HttpTransport.MaxIdleConnsPerHost = 32
	_HttpCli = &http.Client{
		Timeout:   time.Duration(10) * time.Second,
		Transport: &_HttpTransport,
	}
}
