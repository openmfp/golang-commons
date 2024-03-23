package context

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAddSpiffeToContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = AddSpiffeToContext(ctx, "spiffe")

	spiffe, err := GetSpiffeFromContext(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "spiffe", spiffe)
}

func TestAddTenantToContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = AddTenantToContext(ctx, "tenant")

	tenant, err := GetTenantFromContext(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "tenant", tenant)
}

func TestAddAuthHeaderToContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = AddAuthHeaderToContext(ctx, "auth")

	auth, err := GetAuthHeaderFromContext(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "auth", auth)
}

func TestAddWebTokenToContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	issuer := "my-issuer"
	tokenString, err := generateJWT(issuer)
	assert.NoError(t, err)

	ctx = AddWebTokenToContext(ctx, tokenString)

	token, err := GetWebTokenFromContext(ctx)
	assert.Nil(t, err)
	assert.Equal(t, issuer, token.Issuer)
}

func generateJWT(issuer string) (string, error) {
	claims := &jwt.RegisteredClaims{
		Issuer: issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("a_secret_key"))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}
