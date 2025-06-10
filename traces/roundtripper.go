package traces

import (
	"net/http"

	"github.com/openmfp/golang-commons/context/keys"
)

type TracingRoundTripper struct {
	Base http.RoundTripper
}

func (t *TracingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	tracingHeaders, ok := req.Context().Value(keys.TracingHeadersCtxKey).(map[string]string)
	if ok {
		for k, v := range tracingHeaders {
			req.Header.Add(k, v)
		}
	}
	return t.Base.RoundTrip(req)
}
