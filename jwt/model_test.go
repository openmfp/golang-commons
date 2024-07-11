package jwt

import (
	"testing"

	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

var signatureAlgorithms = []jose.SignatureAlgorithm{jose.HS256}

func TestNew(t *testing.T) {
	issuer := "my-issuer"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Issuer: issuer,
	})
	tokenString, err := token.SignedString([]byte("a_secret_key"))
	assert.NoError(t, err)

	webToken, err := New(tokenString, signatureAlgorithms)
	assert.NoError(t, err)
	assert.NotNil(t, webToken)
	assert.Equal(t, issuer, webToken.Issuer)
}

func TestNewAndFail(t *testing.T) {
	tokenString := "just a string"
	_, err := New(tokenString, signatureAlgorithms)
	assert.Error(t, err)

}
