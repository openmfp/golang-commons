package client

import (
	"context"
	"github.com/jellydator/ttlcache/v3"
	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"github.com/openmfp/golang-commons/fga/client/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpenFGAClient_Check(t *testing.T) {
	tenantId := "tenant123"
	storeId := "store123"
	modelId := "model123"
	object := "object"
	relation := "relation"
	user := "user"

	tests := []struct {
		name            string
		prepareCache    func(client *OpenFGAClient)
		clientCheckMock func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient)
		expectedErr     error
	}{
		{
			name: "Check_OK",
			prepareCache: func(client *OpenFGAClient) {
				client.cache.Set(cacheKeyForStore(tenantId), storeId, ttlcache.DefaultTTL)
				client.cache.Set(cacheKeyForModel(tenantId), modelId, ttlcache.DefaultTTL)
			},
			clientCheckMock: func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient) {
				openFGAServiceClientMock.On("Check", ctx, &openfgav1.CheckRequest{
					StoreId:              storeId,
					AuthorizationModelId: modelId,
					TupleKey: &openfgav1.CheckRequestTupleKey{
						Object:   object,
						Relation: relation,
						User:     user,
					},
				}).
					Return(&openfgav1.CheckResponse{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			openFGAServiceClientMock := &mocks.OpenFGAServiceClient{}

			client, err := NewOpenFGAClient(openFGAServiceClientMock)
			assert.NoError(t, err)

			if tt.prepareCache != nil {
				tt.prepareCache(client)
			}

			if tt.clientCheckMock != nil {
				tt.clientCheckMock(ctx, openFGAServiceClientMock)
			}

			_, err = client.Check(ctx, object, relation, user, tenantId)
			assert.Equal(t, tt.expectedErr, err)

			openFGAServiceClientMock.AssertExpectations(t)
		})
	}
}
