package client

import (
	"time"

	"github.com/jellydator/ttlcache/v3"
	openfgav1 "github.com/openfga/api/proto/openfga/v1"

	"github.com/openmfp/golang-commons/fga"
)

var _ fga.OpenFGAClientServicer = (*OpenFGAClient)(nil)

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
