package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacheKeyForModel(t *testing.T) {
	tenantId := "tenant123"
	expected := "modelid-tenant123"
	result := cacheKeyForModel(tenantId)
	assert.Equal(t, expected, result, "they should be equal")
}

func TestCacheKeyForStore(t *testing.T) {
	tenantId := "tenant123"
	expected := "storeid-tenant123"
	result := cacheKeyForStore(tenantId)
	assert.Equal(t, expected, result, "they should be equal")
}
