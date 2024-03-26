package jwt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAudiences(t *testing.T) {
	rawAudiences := []interface{}{
		"audience1",
		"audience2",
		1812, // wrong audience
	}

	token := rawWebToken{
		rawClaims: rawClaims{
			RawAudiences: rawAudiences,
		},
	}

	parsedAudiences := token.getAudiences()

	assert.Contains(t, parsedAudiences, "audience1")
	assert.Contains(t, parsedAudiences, "audience2")
	assert.NotContains(t, parsedAudiences, 1812)
}
func TestParseAudiencesString(t *testing.T) {
	token := rawWebToken{
		rawClaims: rawClaims{
			RawAudiences: "audience1",
		},
	}

	parsedAudiences := token.getAudiences()

	assert.Contains(t, parsedAudiences, "audience1")
	assert.Equal(t, len(parsedAudiences), 1)
}
