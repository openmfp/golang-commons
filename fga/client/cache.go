package client

import "fmt"

func cacheKeyForModel(tenantId string) string {
	return fmt.Sprintf("modelid-%s", tenantId)
}

func cacheKeyForStore(tenantId string) string {
	return fmt.Sprintf("storeid-%s", tenantId)
}
