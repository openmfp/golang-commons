package client

import (
	"context"
	"github.com/jellydator/ttlcache/v3"
	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"github.com/openmfp/golang-commons/fga/client/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpenFGAClient_Write(t *testing.T) {
	tenantId := "tenant123"
	storeId := "store123"
	modelId := "model123"

	tests := []struct {
		name             string
		prepareCache     func(client *OpenFGAClient)
		clientWritesMock func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient)
		expectedErr      error
	}{
		{
			name: "Write_OK",
			prepareCache: func(client *OpenFGAClient) {
				client.cache.Set(cacheKeyForModel(tenantId), modelId, ttlcache.DefaultTTL)
				client.cache.Set(cacheKeyForStore(tenantId), storeId, ttlcache.DefaultTTL)
			},
			clientWritesMock: func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient) {
				openFGAServiceClientMock.
					On("Write", ctx, &openfgav1.WriteRequest{
						StoreId:              storeId,
						AuthorizationModelId: modelId,
						Writes: &openfgav1.WriteRequestWrites{
							TupleKeys: []*openfgav1.TupleKey{
								{
									Object:   "object",
									Relation: "relation",
									User:     "user",
								},
							},
						},
					}).
					Return(nil, nil)
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

			if tt.clientWritesMock != nil {
				tt.clientWritesMock(ctx, openFGAServiceClientMock)
			}

			_, err = client.Write(ctx, "object", "relation", "user", tenantId)
			assert.Equal(t, tt.expectedErr, err)

			openFGAServiceClientMock.AssertExpectations(t)
		})
	}
}

func TestOpenFGAClient_Delete(t *testing.T) {
	tenantId := "tenant123"
	storeId := "store123"
	modelId := "model123"

	tests := []struct {
		name             string
		prepareCache     func(client *OpenFGAClient)
		clientWritesMock func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient)
		expectedErr      error
	}{
		{
			name: "Write_OK",
			prepareCache: func(client *OpenFGAClient) {
				client.cache.Set(cacheKeyForModel(tenantId), modelId, ttlcache.DefaultTTL)
				client.cache.Set(cacheKeyForStore(tenantId), storeId, ttlcache.DefaultTTL)
			},
			clientWritesMock: func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient) {
				openFGAServiceClientMock.
					On("Write", ctx, &openfgav1.WriteRequest{
						StoreId:              storeId,
						AuthorizationModelId: modelId,
						Deletes: &openfgav1.WriteRequestDeletes{
							TupleKeys: []*openfgav1.TupleKeyWithoutCondition{
								{
									Object:   "object",
									Relation: "relation",
									User:     "user",
								},
							},
						},
					}).
					Return(nil, nil)
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

			if tt.clientWritesMock != nil {
				tt.clientWritesMock(ctx, openFGAServiceClientMock)
			}

			_, err = client.Delete(ctx, "object", "relation", "user", tenantId)
			assert.Equal(t, tt.expectedErr, err)

			openFGAServiceClientMock.AssertExpectations(t)
		})
	}
}

func TestOpenFGAClient_WriteIfNeeded(t *testing.T) {
	tenantId := "tenant123"
	storeId := "store123"
	modelId := "model123"
	object := "object"
	relation := "relation"
	user := "user"

	tests := []struct {
		name             string
		prepareCache     func(client *OpenFGAClient)
		clientReadMock   func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient)
		clientWritesMock func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient)
		expectedErr      error
	}{
		{
			name: "WriteIfNeeded_OK",
			prepareCache: func(client *OpenFGAClient) {
				client.cache.Set(cacheKeyForModel(tenantId), modelId, ttlcache.DefaultTTL)
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
			clientWritesMock: func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient) {
				openFGAServiceClientMock.
					On("Write", ctx, &openfgav1.WriteRequest{
						StoreId:              storeId,
						AuthorizationModelId: modelId,
						Writes: &openfgav1.WriteRequestWrites{
							TupleKeys: []*openfgav1.TupleKey{
								{
									Object:   object,
									Relation: relation,
									User:     user,
								},
							},
						},
					}).
					Return(nil, nil)
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

			if tt.clientReadMock != nil {
				tt.clientReadMock(ctx, openFGAServiceClientMock)
			}

			if tt.clientWritesMock != nil {
				tt.clientWritesMock(ctx, openFGAServiceClientMock)
			}

			err = client.WriteIfNeeded(ctx, []*openfgav1.TupleKeyWithoutCondition{
				{
					Object:   object,
					Relation: relation,
					User:     user,
				},
			}, tenantId)
			assert.Equal(t, tt.expectedErr, err)

			openFGAServiceClientMock.AssertExpectations(t)
		})
	}
}

func TestOpenFGAClient_DeleteIfNeeded(t *testing.T) {
	tenantId := "tenant123"
	storeId := "store123"
	modelId := "model123"
	object := "object"
	relation := "relation"
	user := "user"

	tests := []struct {
		name             string
		prepareCache     func(client *OpenFGAClient)
		clientReadMock   func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient)
		clientWritesMock func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient)
		expectedErr      error
	}{
		{
			name: "DeleteIfNeeded_OK",
			prepareCache: func(client *OpenFGAClient) {
				client.cache.Set(cacheKeyForModel(tenantId), modelId, ttlcache.DefaultTTL)
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
			clientWritesMock: func(ctx context.Context, openFGAServiceClientMock *mocks.OpenFGAServiceClient) {
				openFGAServiceClientMock.
					On("Write", ctx, &openfgav1.WriteRequest{
						StoreId:              storeId,
						AuthorizationModelId: modelId,
						Deletes: &openfgav1.WriteRequestDeletes{
							TupleKeys: []*openfgav1.TupleKeyWithoutCondition{
								{
									Object:   "object",
									Relation: "relation",
									User:     "user",
								},
							},
						},
					}).
					Return(nil, nil)
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

			if tt.clientReadMock != nil {
				tt.clientReadMock(ctx, openFGAServiceClientMock)
			}

			if tt.clientWritesMock != nil {
				tt.clientWritesMock(ctx, openFGAServiceClientMock)
			}

			err = client.DeleteIfNeeded(ctx, []*openfgav1.TupleKeyWithoutCondition{
				{
					Object:   object,
					Relation: relation,
					User:     user,
				},
			}, tenantId)
			assert.Equal(t, tt.expectedErr, err)

			openFGAServiceClientMock.AssertExpectations(t)
		})
	}
}
