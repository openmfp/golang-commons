package policy_services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-http-utils/headers"
	"github.com/machinebox/graphql"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	openmfpcontext "github.com/openmfp/golang-commons/context"
	"github.com/openmfp/golang-commons/jwt"
	"github.com/openmfp/golang-commons/logger"
)

func createClient(ctx context.Context, iamApiUrl string) *graphql.Client {
	log := logger.LoadLoggerFromContext(ctx)

	hc := &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	client := graphql.NewClient(iamApiUrl, graphql.WithHTTPClient(hc))

	if log != nil {
		client.Log = log.ComponentLogger("graphql").Trace().Msg
	}
	return client
}

func run(ctx context.Context, client GraphqlClient, request *graphql.Request, resp interface{}, timeout time.Duration) error {
	auth, err := openmfpcontext.GetAuthHeaderFromContext(ctx)
	if err != nil || len(auth) == 0 {
		return fmt.Errorf("the request context does not contain an auth header under the key %q. You can use authz.context to set it", jwt.AuthHeaderCtxKey)
	}
	request.Header.Add(headers.Authorization, auth)
	requestCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return client.Run(requestCtx, request, resp)
}
