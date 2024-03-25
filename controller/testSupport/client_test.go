package testSupport

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFakeClient(t *testing.T) {
	fakeClient := CreateFakeClient(t, &TestApiObject{})

	assert.NotNil(t, fakeClient)
}
