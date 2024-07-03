package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/jellydator/ttlcache/v3"
	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"github.com/openmfp/golang-commons/fga/client/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpenFGAClient_ModelId(t *testing.T) {
	tenantId := "tenant123"
	storeId := "store123"
	modelId := "model123"

	tests := []struct {
		name                       string
		prepareCache               func(client *OpenFGAClient)
		listStoresMock             func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient)
		readAuthorizationModelMock func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient)
		expectedModelId            string
		expectedErr                error
	}{
		{
			name: "ListStores_OK_ReadAuthorizationModels_OK",
			listStoresMock: func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient) {
				openFGAServiceClientMock.
					On("ListStores", ctx, &openfgav1.ListStoresRequest{}).
					Return(&openfgav1.ListStoresResponse{
						Stores: []*openfgav1.Store{
							{
								Id:   "store123",
								Name: fmt.Sprintf("tenant-%s", tenantId),
							},
						},
					}, nil)
			},

			readAuthorizationModelMock: func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient) {
				openFGAServiceClientMock.
					On("ReadAuthorizationModels", ctx, &openfgav1.ReadAuthorizationModelsRequest{StoreId: storeId}).
					Return(&openfgav1.ReadAuthorizationModelsResponse{
						AuthorizationModels: []*openfgav1.AuthorizationModel{
							{
								Id: "model123",
							},
						},
					}, nil)
			},
			expectedModelId: "model123",
		},
		{
			name: "HitModelIdCache_OK",
			prepareCache: func(client *OpenFGAClient) {
				client.cache.Set(cacheKeyForModel(tenantId), modelId, ttlcache.DefaultTTL)
			},
			expectedModelId: "model123",
		},
		{
			name: "HitStoreIdCache_OK",
			prepareCache: func(client *OpenFGAClient) {
				client.cache.Set(cacheKeyForStore(tenantId), storeId, ttlcache.DefaultTTL)
			},
			listStoresMock: nil,
			readAuthorizationModelMock: func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient) {
				openFGAServiceClientMock.
					On("ReadAuthorizationModels", ctx, &openfgav1.ReadAuthorizationModelsRequest{StoreId: storeId}).
					Return(&openfgav1.ReadAuthorizationModelsResponse{
						AuthorizationModels: []*openfgav1.AuthorizationModel{
							{
								Id: "model123",
							},
						},
					}, nil)
			},
			expectedModelId: "model123",
		},
		{
			name: "ListStores_Error",
			listStoresMock: func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient) {
				openFGAServiceClientMock.
					On("ListStores", ctx, &openfgav1.ListStoresRequest{}).
					Return(nil, errors.New("ListStoresError"))
			},
			expectedErr: errors.New("ListStoresError"),
		},
		{
			name: "ReadAuthorizationModels_Error",
			prepareCache: func(client *OpenFGAClient) {
				client.cache.Set(cacheKeyForStore(tenantId), storeId, ttlcache.DefaultTTL)
			},
			listStoresMock: nil,
			readAuthorizationModelMock: func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient) {
				openFGAServiceClientMock.
					On("ReadAuthorizationModels", ctx, &openfgav1.ReadAuthorizationModelsRequest{StoreId: storeId}).
					Return(nil, errors.New("ReadAuthorizationModels"))
			},
			expectedErr: errors.New("ReadAuthorizationModels"),
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

			if tt.listStoresMock != nil {
				tt.listStoresMock(ctx, openFGAServiceClientMock)
			}

			if tt.readAuthorizationModelMock != nil {
				tt.readAuthorizationModelMock(ctx, openFGAServiceClientMock)
			}

			res, err := client.ModelId(ctx, tenantId)
			assert.Equal(t, tt.expectedModelId, res)
			assert.Equal(t, tt.expectedErr, err)

			openFGAServiceClientMock.AssertExpectations(t)
		})
	}
}
