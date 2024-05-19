package httptools

import (
	"net/http"
	"net/textproto"
	"strings"

	"github.com/zeromicro/go-zero/core/mapping"
)

var headerUnmarshaler = mapping.NewUnmarshaler(headerKey, mapping.WithStringValues(),
	mapping.WithCanonicalKeyFunc(textproto.CanonicalMIMEHeaderKey))

// Parse parses the response.
func Parse(resp *http.Response, val interface{}) error {
	if err := ParseHeaders(resp, val); err != nil {
		return err
	}

	return ParseJsonBody(resp, val)
}

// ParseHeaders parses the rsponse headers.
func ParseHeaders(resp *http.Response, val interface{}) error {
	m := map[string]interface{}{}
	for k, v := range resp.Header {
		if len(v) == 1 {
			m[k] = v[0]
		} else {
			m[k] = v
		}
	}

	return headerUnmarshaler.Unmarshal(m, val)
}

// ParseJsonBody parses the rsponse body, which should be in json content type.
func ParseJsonBody(resp *http.Response, val interface{}) error {
	if withJsonBody(resp) {
		return mapping.UnmarshalJsonReader(resp.Body, val)
	}
	return mapping.UnmarshalJsonMap(nil, val)
}

func withJsonBody(r *http.Response) bool {
	return strings.Contains(r.Header.Get(ContentType), ApplicationJson)
}
