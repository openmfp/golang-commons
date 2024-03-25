package context

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	openmfpjwt "github.com/openmfp/golang-commons/jwt"
)

type astruct struct{}

func TestAddSpiffeToContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = AddSpiffeToContext(ctx, "spiffe")

	spiffe, err := GetSpiffeFromContext(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "spiffe", spiffe)
}

func TestWrongSpiffeToContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	key := ContextKey(openmfpjwt.SpiffeCtxKey)
	ctx = context.WithValue(ctx, key, astruct{})

	_, err := GetSpiffeFromContext(ctx)
	assert.Error(t, err, "someone stored a wrong value in the [spiffe] key with type [context.astruct], expected [string]")
}

func TestAddTenantToContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = AddTenantToContext(ctx, "tenant")

	tenant, err := GetTenantFromContext(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "tenant", tenant)
}

func TestAddTenantToContextNegative(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	key := ContextKey(openmfpjwt.TenantIdCtxKey)
	ctx = context.WithValue(ctx, key, astruct{})

	_, err := GetTenantFromContext(ctx)
	assert.Error(t, err, "someone stored a wrong value in the [tenant] key with type [context.astruct], expected [string]")
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

func TestAddWebTokenToContextNegative(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	key := ContextKey(openmfpjwt.WebTokenCtxKey)
	ctx = context.WithValue(ctx, key, astruct{})

	_, err := GetWebTokenFromContext(ctx)
	assert.Error(t, err, "someone stored a wrong value in the [web_token] key with type [context.astruct], expected [jwt.WebToken]")
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

func TestAddIsTechnicalIssuerToContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = AddIsTechnicalIssuerToContext(ctx)

	isTechnicalIssuer := GetIsTechnicalIssuerFromContext(ctx)
	assert.True(t, isTechnicalIssuer)
}
