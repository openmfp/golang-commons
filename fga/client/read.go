package client

import (
	"context"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
)

func (c *OpenFGAClient) Read(ctx context.Context, object *string, relation *string, user *string, tenantId string) (*openfgav1.ReadResponse, error) {
	storeId, err := c.StoreId(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	tk := &openfgav1.ReadRequestTupleKey{}
	if object != nil {
		tk.Object = *object
	}
	if relation != nil {
		tk.Relation = *relation
	}
	if user != nil {
		tk.User = *user
	}

	return c.client.Read(ctx, &openfgav1.ReadRequest{
		StoreId:  storeId,
		TupleKey: tk,
	})
}

func (c *OpenFGAClient) Exists(ctx context.Context, tuple *openfgav1.TupleKeyWithoutCondition, tenantId string) (bool, error) {
	resp, err := c.Read(ctx, &tuple.Object, &tuple.Relation, &tuple.User, tenantId)
	if err != nil {
		return false, err
	}
	if len(resp.Tuples) == 0 {
		return false, nil
	}
	return true, nil
}
