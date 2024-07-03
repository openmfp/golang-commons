package helpers

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
		setupMock       func(client *mocks.OpenFGAServiceClient)
		expectedModelID string
		expectedError   error
	}{
		{
			name: "FullPath_OK",
			setupMock: func(client *mocks.OpenFGAServiceClient) {
				client.On("ListStores", ctx, &openfgav1.ListStoresRequest{}).Return(&openfgav1.ListStoresResponse{
					Stores: []*openfgav1.Store{
						{Id: "store123", Name: "tenant-tenant123"},
					},
				}, nil)
				client.On("ReadAuthorizationModels", ctx, &openfgav1.ReadAuthorizationModelsRequest{StoreId: "store123"}).Return(&openfgav1.ReadAuthorizationModelsResponse{
					AuthorizationModels: []*openfgav1.AuthorizationModel{
						{Id: modelId},
					},
				}, nil)
			},
			expectedModelID: modelId,
			expectedError:   nil,
		},
		{
			name: "HitGetModelIDForTenantCache_OK",
			setupMock: func(client *mocks.OpenFGAServiceClient) {
				cache.Add("model-tenant123", modelId)
			},
			expectedModelID: modelId,
			expectedError:   nil,
		},
		{
			name: "HitGetStoreIDForTenantCache_OK",
			setupMock: func(client *mocks.OpenFGAServiceClient) {
				cache.Add("tenant-tenant123", storeId)
				client.On("ReadAuthorizationModels", ctx, &openfgav1.ReadAuthorizationModelsRequest{StoreId: "store123"}).
					Return(&openfgav1.ReadAuthorizationModelsResponse{
						AuthorizationModels: []*openfgav1.AuthorizationModel{
							{Id: modelId},
						},
					}, nil)
			},
			expectedModelID: modelId,
			expectedError:   nil,
		},
		{
			name: "ListStores_Error",
			setupMock: func(client *mocks.OpenFGAServiceClient) {
				client.On("ListStores", ctx, &openfgav1.ListStoresRequest{}).
					Return(nil, assert.AnError)
			},
			expectedError: assert.AnError,
		},
		{
			name: "MatchingKeyNotFound_Error",
			setupMock: func(client *mocks.OpenFGAServiceClient) {
				client.On("ListStores", ctx, &openfgav1.ListStoresRequest{}).Return(&openfgav1.ListStoresResponse{
					Stores: []*openfgav1.Store{}}, nil)
			},
			expectedError: errors.New("could not find store matching key \"tenant-tenant123\""),
		},
		{
			name: "ReadAuthorizationModels_Error",
			setupMock: func(client *mocks.OpenFGAServiceClient) {
				cache.Add("tenant-tenant123", storeId)

				client.On("ReadAuthorizationModels", ctx, &openfgav1.ReadAuthorizationModelsRequest{
					StoreId: "store123"},
				).Return(nil, assert.AnError)
			},
			expectedError: assert.AnError,
		},
		{
			name: "NoReadAuthorizationModels_Error",
			setupMock: func(client *mocks.OpenFGAServiceClient) {
				cache.Add("tenant-tenant123", storeId)

				client.On("ReadAuthorizationModels", ctx, &openfgav1.ReadAuthorizationModelsRequest{StoreId: "store123"}).
					Return(&openfgav1.ReadAuthorizationModelsResponse{}, nil)
			},
			expectedError: errors.New("no authorization models in response. Cannot determine proper AuthorizationModelId"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache.Purge() // Clear cache before each test

			client := &mocks.OpenFGAServiceClient{}
			tt.setupMock(client)

			modelID, err := GetModelIDForTenant(ctx, client, tenantID)

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
		//{
		//	name:     "DuplicateWriteGRPCError",
		//	err:      status.Error(codes.InvalidArgument, openfgav1.ErrorCode_write_failed_due_to_invalid_input.String()),
		//	expected: true,
		//},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDuplicateWriteError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
