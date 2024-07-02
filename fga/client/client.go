package client

import (
	"context"
	"time"

	"github.com/jellydator/ttlcache/v3"
	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	Dispose()
}

var _ OpenFGAClientServicer = (*OpenFGAClient)(nil)

type OpenFGAClient struct {
	client openfgav1.OpenFGAServiceClient
	conn   *grpc.ClientConn
	cache  *ttlcache.Cache[string, string]
}

func NewOpenFGAClient(host string) (*OpenFGAClient, error) {

	conn, err := grpc.Dial(host,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	cache := ttlcache.New[string, string](
		ttlcache.WithTTL[string, string](5 * time.Minute),
	)

	go cache.Start()
	grpClient := openfgav1.NewOpenFGAServiceClient(conn)
	return &OpenFGAClient{
		client: grpClient,
		conn:   conn,
		cache:  cache,
	}, nil
}

func (c *OpenFGAClient) Dispose() {
	c.conn.Close()
}
