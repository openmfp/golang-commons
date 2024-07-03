package client

import (
	"context"
	"time"

	"github.com/jellydator/ttlcache/v3"
	openfgav1 "github.com/openfga/api/proto/openfga/v1"
)

//go:generate mockery --name OpenFGAClientServicer --output ./mocks --filename client.go
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

var _ OpenFGAClientServicer = (*OpenFGAClient)(nil)

type OpenFGAClient struct {
	client openfgav1.OpenFGAServiceClient
	cache  *ttlcache.Cache[string, string]
}

func NewOpenFGAClient(openFGAServiceClient openfgav1.OpenFGAServiceClient) (*OpenFGAClient, error) {
	cache := ttlcache.New[string, string](
		ttlcache.WithTTL[string, string](5 * time.Minute),
	)

	go cache.Start()

	return &OpenFGAClient{
		client: openFGAServiceClient,
		cache:  cache,
	}, nil
}
