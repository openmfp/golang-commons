package traces

import (
	"context"
	"net/http"

	"github.com/openmfp/golang-commons/context/keys"
)

var availableTracingHeaders = []string{
	"X-Request-Id",
	"X-B3-Traceid",
	"X-B3-Spanid",
	"X-B3-Parentspanid",
	"X-B3-Sampled",
	"X-B3-Flags",
	"X-Ot-Span-Context",
	"Traceparent",
}

func SetTracingHeadersInContext() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {

			tracingHeaders := map[string]string{}
			for _, key := range availableTracingHeaders {
				val := request.Header.Get(key)
				if len(val) > 0 {
					tracingHeaders[key] = val
				}
			}
			ctx := context.WithValue(request.Context(), keys.TracingHeadersCtxKey, tracingHeaders)
			next.ServeHTTP(responseWriter, request.WithContext(ctx))
		})
	}
}
