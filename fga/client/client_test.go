package client

import (
	"github.com/openmfp/golang-commons/fga/client/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewOpenFGAClient(t *testing.T) {
	client, err := NewOpenFGAClient(&mocks.OpenFGAServiceClient{})
	assert.NoError(t, err)
	assert.NotNil(t, client)
}
