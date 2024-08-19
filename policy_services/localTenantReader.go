package policy_services

import (
	"context"
)

type localTenantReader struct {
	tenantId string
}

// NewLocalTenantRetriever Creates a Tenant reader that returns a hardcoded tenant id for local testing.
// The idea is to use a tenant id that can be set to the environment, so you do not need an iam service running locally
func NewLocalTenantRetriever(tenantId string) *TenantRetrieverService {
	tr := &localTenantReader{
		tenantId: tenantId,
	}
	return NewCustomTenantRetriever(tr)
}

func (reader localTenantReader) Read(context.Context) (string, error) {
	return reader.tenantId, nil
}
