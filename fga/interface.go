package fga

import (
	"context"
	"github.com/openfga/api/proto/openfga/v1"
)

type OpenFGAClientServicer interface {
	Check(ctx context.Context, object string, relation string, user string, tenantId string) (*openfgav1.CheckResponse, error)
	Read(ctx context.Context, object *string, relation *string, user *string, tenantId string) (*openfgav1.ReadResponse, error)
	Exists(ctx context.Context, tuple *openfgav1.TupleKeyWithoutCondition, tenantId string) (bool, error)
	Writes(ctx context.Context, writes []*openfgav1.TupleKey, deletes []*openfgav1.TupleKeyWithoutCondition, tenantId string) (bool, error)
	Write(ctx context.Context, object string, relation string, user string, tenantId string) (bool, error)
	WriteIfNeeded(ctx context.Context, tuples []*openfgav1.TupleKeyWithoutCondition, tenantId string) error
	DeleteIfNeeded(ctx context.Context, tuples []*openfgav1.TupleKeyWithoutCondition, tenantId string) error
	Delete(ctx context.Context, object string, relation string, user string, tenantId string) (bool, error)
	ModelId(ctx context.Context, tenantId string) (string, error)
	StoreId(ctx context.Context, tenantId string) (string, error)
}
