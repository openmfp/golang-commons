package store

import (
	"context"
	"errors"
	"testing"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/openmfp/golang-commons/fga/client/mocks"
)

func TestGetModelIDForTenant(t *testing.T) {
	ctx := context.Background()
	tenantID := "tenant123"
	storeId := "store123"
	modelId := "model123"

	tests := []struct {
		name            string
		setupMock       func(client *mocks.OpenFGAServiceClient, cachedStore *Service)
		expectedModelID string
		expectedError   error
	}{
		{
			name: "FullPath_OK",
			setupMock: func(client *mocks.OpenFGAServiceClient, cachedStore *Service) {
				client.EXPECT().
					ListStores(ctx, &openfgav1.ListStoresRequest{}).
					Return(&openfgav1.ListStoresResponse{
						Stores: []*openfgav1.Store{
							{Id: "store123", Name: "tenant-tenant123"},
						}}, nil).
					Once()

				client.EXPECT().
					ReadAuthorizationModels(ctx, &openfgav1.ReadAuthorizationModelsRequest{StoreId: "store123"}).
					Return(&openfgav1.ReadAuthorizationModelsResponse{
						AuthorizationModels: []*openfgav1.AuthorizationModel{
							{Id: modelId},
						}}, nil).
					Once()
			},
			expectedModelID: modelId,
			expectedError:   nil,
		},
		{
			name: "HitGetModelIDForTenantCache_OK",
			setupMock: func(client *mocks.OpenFGAServiceClient, cachedStore *Service) {
				cachedStore.GetCache().Add("model-tenant123", modelId)
			},
			expectedModelID: modelId,
			expectedError:   nil,
		},
		{
			name: "HitGetStoreIDForTenantCache_OK",
			setupMock: func(client *mocks.OpenFGAServiceClient, cachedStore *Service) {
				cachedStore.GetCache().Add("tenant-tenant123", storeId)

				client.EXPECT().
					ReadAuthorizationModels(ctx, &openfgav1.ReadAuthorizationModelsRequest{StoreId: "store123"}).
					Return(&openfgav1.ReadAuthorizationModelsResponse{
						AuthorizationModels: []*openfgav1.AuthorizationModel{
							{Id: modelId},
						}}, nil).
					Once()
			},
			expectedModelID: modelId,
			expectedError:   nil,
		},
		{
			name: "ListStores_Error",
			setupMock: func(client *mocks.OpenFGAServiceClient, cachedStore *Service) {
				client.EXPECT().
					ListStores(ctx, &openfgav1.ListStoresRequest{}).
					Return(nil, assert.AnError).
					Once()
			},
			expectedError: assert.AnError,
		},
		{
			name: "MatchingKeyNotFound_Error",
			setupMock: func(client *mocks.OpenFGAServiceClient, cachedStore *Service) {
				client.EXPECT().
					ListStores(ctx, &openfgav1.ListStoresRequest{}).
					Return(&openfgav1.ListStoresResponse{Stores: []*openfgav1.Store{}}, nil).
					Once()
			},
			expectedError: errors.New("could not find store matching key \"tenant-tenant123\""),
		},
		{
			name: "ReadAuthorizationModels_Error",
			setupMock: func(client *mocks.OpenFGAServiceClient, cachedStore *Service) {
				cachedStore.GetCache().Add("tenant-tenant123", storeId)

				client.EXPECT().
					ReadAuthorizationModels(ctx, &openfgav1.ReadAuthorizationModelsRequest{StoreId: "store123"}).
					Return(nil, assert.AnError).
					Once()
			},
			expectedError: assert.AnError,
		},
		{
			name: "NoReadAuthorizationModels_Error",
			setupMock: func(client *mocks.OpenFGAServiceClient, cachedStore *Service) {
				cachedStore.GetCache().Add("tenant-tenant123", storeId)

				client.EXPECT().
					ReadAuthorizationModels(ctx, &openfgav1.ReadAuthorizationModelsRequest{StoreId: "store123"}).
					Return(&openfgav1.ReadAuthorizationModelsResponse{}, nil).
					Once()
			},
			expectedError: errors.New("no authorization models in response. Cannot determine proper AuthorizationModelId"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mocks.OpenFGAServiceClient{}
			cachedStore := New()
			tt.setupMock(client, cachedStore)

			modelID, err := cachedStore.GetModelIDForTenant(ctx, client, tenantID)

			assert.Equal(t, tt.expectedModelID, modelID)
			assert.Equal(t, tt.expectedError, err)

			client.AssertExpectations(t)
		})
	}
}

func TestIsDuplicateWriteError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "NoError",
			err:      nil,
			expected: false,
		},
		{
			name:     "NonGRPCError",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "NonDuplicateWriteGRPCError",
			err:      status.Error(codes.InvalidArgument, "invalid argument"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cachedStore := New()

			result := cachedStore.IsDuplicateWriteError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
