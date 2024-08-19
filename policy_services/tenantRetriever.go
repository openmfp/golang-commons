package policy_services

import (
	"context"
	"fmt"
	"strings"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/machinebox/graphql"
	openmfpcontext "github.com/openmfp/golang-commons/context"
	"github.com/openmfp/golang-commons/logger"
)

type GraphqlData struct {
	TenantInfo TenantResponse
}

type TenantResponse struct {
	TenantId string
}

type TenantKey string

type TenantRetrieverService struct {
	tenantReader TenantIdReader
	tenantCache  *lru.Cache[TenantKey, string]
}

type GraphqlClient interface {
	Run(ctx context.Context, req *graphql.Request, resp interface{}) error
}

type TenantIdReader interface {
	Read(parentCtx context.Context) (string, error)
}

// NewCustomTenantRetriever Creates a tenant retriever that has a custom implementation
// of the TenantIdReader interface. The result of the reader will be cached.
// A local and a graphql tenant Reader are already contained in this package.
func NewCustomTenantRetriever(tr TenantIdReader) *TenantRetrieverService {
	cache, err := lru.New[TenantKey, string](5)
	if err != nil {
		logger.StdLogger.Error().Err(err).Msg("cannot initialize cache, working without memory")
	}

	return &TenantRetrieverService{
		tenantReader: tr,
		tenantCache:  cache,
	}
}

// TenantRetriever Retrieves a tenant by a issuer/audiences from the context. It uses the iam service's graphql endpoint.
// The responses are cached in memory because tenant mappings do not change frequently
// It uses the tenantInfo query internally.
type TenantRetriever interface {
	RetrieveTenant(ctx context.Context) (string, error)
}

func (tenantRetriever *TenantRetrieverService) RetrieveTenant(ctx context.Context) (string, error) {
	webToken, err := openmfpcontext.GetWebTokenFromContext(ctx)
	if err != nil {
		return "", nil
	}
	if webToken.Issuer == "" {
		return "", nil
	}
	if len(webToken.Audiences) == 0 {
		return "", nil
	}

	key := TenantKey(fmt.Sprintf("%s-%s", webToken.Issuer, strings.Join(webToken.Audiences, "-")))
	return tenantRetriever.RetrieveOrAdd(ctx, key)
}

func (tenantRetriever *TenantRetrieverService) RetrieveOrAdd(ctx context.Context, key TenantKey) (string, error) {
	cachedTenant, found := tenantRetriever.tenantCache.Get(key)
	if found {
		return cachedTenant, nil
	}

	tenantId, err := tenantRetriever.tenantReader.Read(ctx)
	if err != nil {
		return tenantId, err
	}

	tenantRetriever.tenantCache.Add(key, tenantId)
	return tenantId, nil
}
