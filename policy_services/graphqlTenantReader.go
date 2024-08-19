package policy_services

import (
	"context"
	"fmt"
	"time"

	"github.com/machinebox/graphql"
)

type graphqlTenantReader struct {
	client  GraphqlClient
	iamUrl  string
	timeout time.Duration
}

// NewTenantRetriever Creates a retriever to get tenant ids from the iam service.
// The iamUrl parameter should be the graphql endpoint of the iam-service.
func NewTenantRetriever(ctx context.Context, iamUrl string, timeout *time.Duration) *TenantRetrieverService {
	tr := &graphqlTenantReader{
		client:  createClient(ctx, iamUrl),
		iamUrl:  iamUrl,
		timeout: time.Second * 5,
	}
	if timeout != nil {
		tr.timeout = *timeout
	}

	return NewCustomTenantRetriever(tr)
}

func (r *graphqlTenantReader) Read(ctx context.Context) (string, error) {
	req := graphql.NewRequest(`
		  query {
				tenantInfo {
					tenantId
			  }
			}
	`)

	var respData GraphqlData
	if err := run(ctx, r.client, req, &respData, r.timeout); err != nil {
		return "", err
	}

	id := respData.TenantInfo.TenantId

	if id == "" {
		return "", fmt.Errorf("the tenantInfo query returned no tenant id. The iam service %s was called", r.iamUrl)
	}

	return id, nil
}
