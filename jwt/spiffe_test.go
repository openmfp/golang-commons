package jwt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpiffe(t *testing.T) {
	t.Run("GetSpiffeUrlValue", func(t *testing.T) {
		// Arrange
		header := make(map[string][]string)
		header[HeaderSpiffeValue] = []string{"URI=spiffe://example.com"}

		// Act
		got := GetSpiffeUrlValue(header)

		// Assert
		assert.NotNil(t, got)
		assert.Equal(t, "spiffe://example.com", *got)
	})

	t.Run("GetURIValueForNotMatchingRegEx", func(t *testing.T) {
		// Act
		got := GetURIValue("spiffe://example.com")

		// Assert
		assert.Equal(t, "", got)
	})
}
