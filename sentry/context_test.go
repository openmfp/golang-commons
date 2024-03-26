package sentry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextWithSentryTags(t *testing.T) {
	ctx := context.Background()
	tags := map[string]string{
		"key": "value",
	}

	ctx = ContextWithSentryTags(ctx, tags)
	assert.Equal(t, tags, GetSentryTagsFromContext(ctx))
}
