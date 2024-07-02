package client

import (
	"context"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
)

func (c *OpenFGAClient) Check(ctx context.Context, object string, relation string, user string, tenantId string) (*openfgav1.CheckResponse, error) {
	storeId, err := c.StoreId(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	modelId, err := c.ModelId(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	return c.client.Check(ctx, &openfgav1.CheckRequest{
		StoreId:              storeId,
		AuthorizationModelId: modelId,
		TupleKey: &openfgav1.CheckRequestTupleKey{
			Object:   object,
			Relation: relation,
			User:     user,
		},
	})
}
