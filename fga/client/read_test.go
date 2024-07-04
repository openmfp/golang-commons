package client

import (
	"context"
	"github.com/jellydator/ttlcache/v3"
	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"github.com/openmfp/golang-commons/fga/client/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpenFGAClient_Exists(t *testing.T) {
	tenantId := "tenant123"
	storeId := "store123"
	object := "object"
	relation := "relation"
	user := "user"

	tests := []struct {
		name             string
		setupMock        func(ctx context.Context, client *OpenFGAClient, openFGAServiceClientMock *mocks.OpenFGAServiceClient)
		expectedResponse bool
		expectedErr      error
	}{
		{
			name: "Exists_OK",
			setupMock: func(ctx context.Context, client *OpenFGAClient, openFGAServiceClientMock *mocks.OpenFGAServiceClient) {
				client.cache.Set(cacheKeyForStore(tenantId), storeId, ttlcache.DefaultTTL)

				openFGAServiceClientMock.EXPECT().
					Read(ctx, &openfgav1.ReadRequest{
						StoreId: storeId,
						TupleKey: &openfgav1.ReadRequestTupleKey{
							Object:   object,
							Relation: relation,
							User:     user,
						}}).
					Return(&openfgav1.ReadResponse{Tuples: []*openfgav1.Tuple{{Key: nil}}}, nil).
					Once()
			},
			expectedResponse: true,
		},
		{
			name: "Exists_Read_Error",
			setupMock: func(ctx context.Context, client *OpenFGAClient, openFGAServiceClientMock *mocks.OpenFGAServiceClient) {
				client.cache.Set(cacheKeyForStore(tenantId), storeId, ttlcache.DefaultTTL)

				openFGAServiceClientMock.EXPECT().
					Read(ctx, &openfgav1.ReadRequest{
						StoreId: storeId,
						TupleKey: &openfgav1.ReadRequestTupleKey{
							Object:   object,
							Relation: relation,
							User:     user,
						}}).
					Return(nil, assert.AnError).
					Once()
			},
			expectedResponse: false,
			expectedErr:      assert.AnError,
		},
		{
			name: "Exists_No_Tuples_Error",
			setupMock: func(ctx context.Context, client *OpenFGAClient, openFGAServiceClientMock *mocks.OpenFGAServiceClient) {
				client.cache.Set(cacheKeyForStore(tenantId), storeId, ttlcache.DefaultTTL)

				openFGAServiceClientMock.EXPECT().
					Read(ctx, &openfgav1.ReadRequest{
						StoreId: storeId,
						TupleKey: &openfgav1.ReadRequestTupleKey{
							Object:   object,
							Relation: relation,
							User:     user,
						}}).
					Return(&openfgav1.ReadResponse{}, nil).
					Once()
			},
			expectedResponse: false,
			expectedErr:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			openFGAServiceClientMock := &mocks.OpenFGAServiceClient{}

			client, err := NewOpenFGAClient(openFGAServiceClientMock)
			assert.NoError(t, err)

			if tt.setupMock != nil {
				tt.setupMock(ctx, client, openFGAServiceClientMock)
			}

			res, err := client.Exists(ctx, &openfgav1.TupleKeyWithoutCondition{
				Object:   object,
				Relation: relation,
				User:     user,
			}, tenantId)
			assert.Equal(t, tt.expectedResponse, res)
			assert.Equal(t, tt.expectedErr, err)

			openFGAServiceClientMock.AssertExpectations(t)
		})
	}
}
