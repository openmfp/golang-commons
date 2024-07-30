package store

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"google.golang.org/grpc/status"
)

type FgaTenantStore interface {
	GetStoreIDForTenant(ctx context.Context, conn openfgav1.OpenFGAServiceClient, tenantID string) (string, error)
	GetModelIDForTenant(ctx context.Context, conn openfgav1.OpenFGAServiceClient, tenantID string) (string, error)
	IsDuplicateWriteError(err error) bool
	GetCache() *expirable.LRU[string, string]
}

type Service struct {
	cache *expirable.LRU[string, string]
	FgaTenantStore
}

func New() *Service {
	return &Service{
		cache: expirable.NewLRU[string, string](10, nil, 10*time.Minute),
	}
}

func (c *Service) GetStoreIDForTenant(ctx context.Context, conn openfgav1.OpenFGAServiceClient, tenantID string) (string, error) {

	cacheKey := "tenant-" + tenantID
	s, ok := c.cache.Get(cacheKey)
	if ok && s != "" {
		return s, nil
	}

	res, err := conn.ListStores(ctx, &openfgav1.ListStoresRequest{})
	if err != nil {
		return "", err
	}

	idx := slices.IndexFunc(res.GetStores(), func(s *openfgav1.Store) bool { return s.Name == cacheKey })
	if idx < 0 {
		return "", fmt.Errorf("could not find store matching key %q", cacheKey)
	}

	store := res.GetStores()[idx]
	c.cache.Add(cacheKey, store.Id)

	return store.Id, nil
}

func (c *Service) GetModelIDForTenant(ctx context.Context, conn openfgav1.OpenFGAServiceClient, tenantID string) (string, error) {

	cacheKey := "model-" + tenantID
	s, ok := c.cache.Get(cacheKey)
	if ok && s != "" {
		return s, nil
	}

	storeID, err := c.GetStoreIDForTenant(ctx, conn, tenantID)
	if err != nil {
		return "", err
	}

	res, err := conn.ReadAuthorizationModels(ctx, &openfgav1.ReadAuthorizationModelsRequest{StoreId: storeID})
	if err != nil {
		return "", err
	}

	if len(res.AuthorizationModels) < 1 {
		return "", errors.New("no authorization models in response. Cannot determine proper AuthorizationModelId")
	}

	modelID := res.AuthorizationModels[0].Id
	c.cache.Add(cacheKey, modelID)

	return modelID, nil
}

func (c *Service) IsDuplicateWriteError(err error) bool {
	if err == nil {
		return false
	}

	s, ok := status.FromError(err)
	return ok && int32(s.Code()) == int32(openfgav1.ErrorCode_write_failed_due_to_invalid_input)
}

func (c *Service) GetCache() *expirable.LRU[string, string] {
	return c.cache
}
