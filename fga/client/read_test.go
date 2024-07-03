package client

import (
	"context"
	"errors"
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
		prepareCache     func(client *OpenFGAClient)
		clientReadMock   func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient)
		expectedResponse bool
		expectedErr      error
	}{
		{
			name: "Exists_OK",
			prepareCache: func(client *OpenFGAClient) {
				client.cache.Set(cacheKeyForStore(tenantId), storeId, ttlcache.DefaultTTL)
			},
			clientReadMock: func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient) {
				openFGAServiceClientMock.On("Read", ctx, &openfgav1.ReadRequest{
					StoreId: storeId,
					TupleKey: &openfgav1.ReadRequestTupleKey{
						Object:   object,
						Relation: relation,
						User:     user,
					},
				}).
					Return(&openfgav1.ReadResponse{
						Tuples: []*openfgav1.Tuple{
							{
								Key: nil,
							},
						},
					}, nil)
			},
			expectedResponse: true,
		},
		{
			name: "Exists_Read_Error",
			prepareCache: func(client *OpenFGAClient) {
				client.cache.Set(cacheKeyForStore(tenantId), storeId, ttlcache.DefaultTTL)
			},
			clientReadMock: func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient) {
				openFGAServiceClientMock.On("Read", ctx, &openfgav1.ReadRequest{
					StoreId: storeId,
					TupleKey: &openfgav1.ReadRequestTupleKey{
						Object:   object,
						Relation: relation,
						User:     user,
					},
				}).
					Return(nil, errors.New("Exists_Read_Error"))
			},
			expectedResponse: false,
			expectedErr:      errors.New("Exists_Read_Error"),
		},
		{
			name: "Exists_No_Tuples_Error",
			prepareCache: func(client *OpenFGAClient) {
				client.cache.Set(cacheKeyForStore(tenantId), storeId, ttlcache.DefaultTTL)
			},
			clientReadMock: func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient) {
				openFGAServiceClientMock.On("Read", ctx, &openfgav1.ReadRequest{
					StoreId: storeId,
					TupleKey: &openfgav1.ReadRequestTupleKey{
						Object:   object,
						Relation: relation,
						User:     user,
					},
				}).
					Return(&openfgav1.ReadResponse{}, nil)
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

			if tt.prepareCache != nil {
				tt.prepareCache(client)
			}

			if tt.clientReadMock != nil {
				tt.clientReadMock(ctx, openFGAServiceClientMock)
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
