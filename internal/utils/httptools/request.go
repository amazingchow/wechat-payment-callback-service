package httptools

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	nurl "net/url"
	"strings"

	"github.com/zeromicro/go-zero/core/mapping"

	"github.com/amazingchow/wechat-payment-callback-service/internal/common/logger"
)

type (
	client interface {
		do(r *http.Request) (*http.Response, error)
	}

	defaultClient struct{}
)

func (c defaultClient) do(r *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(r)
}

func request(r *http.Request, cli client) (*http.Response, error) {
	return cli.do(r)
}

// Do sends an HTTP request with the given arguments and returns an HTTP response,
// data is automatically marshal into a *httpRequest.
func Do(ctx context.Context, method, url string, data interface{}, opentrace bool) (*http.Response, error) {
	req, err := buildRequest(ctx, method, url, data, opentrace)
	if err != nil {
		return nil, err
	}

	return DoRequest(req)
}

// DoRequest sends an HTTP request and returns an HTTP response.
func DoRequest(r *http.Request) (*http.Response, error) {
	return request(r, defaultClient{})
}

// DoWithCustomClient sends an HTTP request with the given arguments and returns an HTTP response,
// data is automatically marshal into a *httpRequest.
func DoWithCustomClient(ctx context.Context, cli *http.Client, method, url string, data interface{}, opentrace bool) (*http.Response, error) {
	req, err := buildRequest(ctx, method, url, data, opentrace)
	if err != nil {
		return nil, err
	}

	return DoRequestWithCustomClient(cli, req)
}

// DoRequestWithCustomClient sends an HTTP request and returns an HTTP response.
func DoRequestWithCustomClient(cli *http.Client, r *http.Request) (*http.Response, error) {
	return cli.Do(r)
}

func buildFormQuery(u *nurl.URL, val map[string]interface{}) string {
	query := u.Query()
	for k, v := range val {
		query.Add(k, fmt.Sprint(v))
	}

	return query.Encode()
}

func buildRequest(ctx context.Context, method, url string, data interface{}, opentrace bool) (*http.Request, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return nil, err
	}

	var val map[string]map[string]interface{}
	if data != nil {
		val, err = mapping.Marshal(data)
		if err != nil {
			return nil, err
		}
	}

	if err := fillPath(u, val[pathKey]); err != nil {
		return nil, err
	}

	var reader io.Reader
	jsonVars, hasJsonBody := val[jsonKey]
	if hasJsonBody {
		if method == http.MethodGet {
			return nil, ErrGetWithBody
		}

		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		if err := enc.Encode(jsonVars); err != nil {
			return nil, err
		}

		reader = &buf
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = buildFormQuery(u, val[formKey])
	fillHeader(req, val[headerKey])
	if hasJsonBody {
		req.Header.Set(ContentType, JsonContentType)
	}

	if opentrace {
		req = req.WithContext(httptrace.WithClientTrace(req.Context(), withClientTrace()))
	}

	return req, nil
}

func fillHeader(r *http.Request, val map[string]interface{}) {
	for k, v := range val {
		r.Header.Add(k, fmt.Sprint(v))
	}
}

func fillPath(u *nurl.URL, val map[string]interface{}) error {
	used := make(map[string]struct{})
	fields := strings.Split(u.Path, slash)

	for i := range fields {
		field := fields[i]
		if len(field) > 0 && field[0] == colon {
			name := field[1:]
			ival, ok := val[name]
			if !ok {
				return fmt.Errorf("missing path variable %q", name)
			}
			value := fmt.Sprint(ival)
			if len(value) == 0 {
				return fmt.Errorf("empty path variable %q", name)
			}
			fields[i] = value
			used[name] = struct{}{}
		}
	}

	if len(val) != len(used) {
		for key := range used {
			delete(val, key)
		}

		var unused []string
		for key := range val {
			unused = append(unused, key)
		}

		return fmt.Errorf("more path variables are provided: %q", strings.Join(unused, ", "))
	}

	u.Path = strings.Join(fields, slash)
	return nil
}

func withClientTrace() *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) {
			logger.GetGlobalLogger().Debugf("[http-tracing] a successful connection was obtained, reused: %v, idle: %v", info.Reused, info.WasIdle)
		},
		PutIdleConn: func(err error) {
			if err == nil {
				logger.GetGlobalLogger().Debug("[http-tracing] the connection was successfully returned to the idle pool")
			} else {
				logger.GetGlobalLogger().WithError(err).Error("[http-tracing] the connection was failed to return to the idle pool")
			}
		},
		GotFirstResponseByte: func() {
			logger.GetGlobalLogger().Debug("[http-tracing] the first byte of the response headers is available")
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			logger.GetGlobalLogger().Debugf("[http-tracing] to find dns lookup for %s", info.Host)
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			if info.Err == nil {
				logger.GetGlobalLogger().Debug("[http-tracing] found dns lookup")
			} else {
				logger.GetGlobalLogger().WithError(info.Err).Error("[http-tracing] failed to find dns lookup")
			}
		},
		ConnectStart: func(network, addr string) {
			logger.GetGlobalLogger().Debugf("[http-tracing] a new connection's Dial begins")
		},
		ConnectDone: func(network, addr string, err error) {
			logger.GetGlobalLogger().Debugf("[http-tracing] a new connection's Dial completes")
		},
		TLSHandshakeStart: func() {
			logger.GetGlobalLogger().Debugf("[http-tracing] the TLS handshake begins")
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			if err == nil {
				logger.GetGlobalLogger().Debugf("[http-tracing] the TLS handshake completes")
			} else {
				logger.GetGlobalLogger().WithError(err).Error("[http-tracing] the TLS handshake was failed")
			}
		},
	}
}
