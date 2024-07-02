package client

import (
	"context"
	"fmt"

	"github.com/jellydator/ttlcache/v3"
	openfgav1 "github.com/openfga/api/proto/openfga/v1"
)

func (c *OpenFGAClient) ModelId(ctx context.Context, tenantId string) (string, error) {
	if cacheItem := c.cache.Get(cacheKeyForModel(tenantId)); cacheItem != nil {
		val := cacheItem.Value()
		return val, nil
	}

	storeId, err := c.StoreId(ctx, tenantId)
	if err != nil {
		return "", err
	}

	resp, err := c.client.ReadAuthorizationModels(ctx, &openfgav1.ReadAuthorizationModelsRequest{StoreId: storeId})
	if err != nil {
		return "", err
	}

	if len(resp.AuthorizationModels) > 0 {
		c.cache.Set(cacheKeyForModel(tenantId), resp.AuthorizationModels[0].Id, ttlcache.DefaultTTL)
		return resp.AuthorizationModels[0].Id, nil
	}

	return "", fmt.Errorf("could not determine store. No stores found")
}

func (c *OpenFGAClient) StoreId(ctx context.Context, tenantId string) (string, error) {
	if cacheItem := c.cache.Get(cacheKeyForStore(tenantId)); cacheItem != nil {
		val := cacheItem.Value()
		return val, nil
	}

	expectedStoreName := fmt.Sprintf("tenant-%s", tenantId)
	resp, err := c.client.ListStores(ctx, &openfgav1.ListStoresRequest{})
	if err != nil {
		return "", err
	}

	for _, store := range resp.Stores {
		if store.Name == expectedStoreName {
			c.cache.Set(cacheKeyForStore(tenantId), store.Id, ttlcache.DefaultTTL)
			return store.Id, nil
		}
	}
	return "", fmt.Errorf("could not determine model. No models found")
}
