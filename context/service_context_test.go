package context_test

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	openmfpctx "github.com/openmfp/golang-commons/context"
	"github.com/openmfp/golang-commons/context/keys"
	openmfpjwt "github.com/openmfp/golang-commons/jwt"
)

type astruct struct{}

func TestAddSpiffeToContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = openmfpctx.AddSpiffeToContext(ctx, "spiffe")

	spiffe, err := openmfpctx.GetSpiffeFromContext(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "spiffe", spiffe)
}

func TestWrongSpiffeToContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	key := openmfpctx.ContextKey(openmfpjwt.SpiffeCtxKey)
	ctx = context.WithValue(ctx, key, astruct{})

	_, err := openmfpctx.GetSpiffeFromContext(ctx)
	assert.Error(t, err, "someone stored a wrong value in the [spiffe] key with type [context.astruct], expected [string]")
}

func TestAddTenantToContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = openmfpctx.AddTenantToContext(ctx, "tenant")

	tenant, err := openmfpctx.GetTenantFromContext(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "tenant", tenant)
}

func TestAddTenantToContextNegative(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	key := openmfpctx.ContextKey(openmfpjwt.TenantIdCtxKey)
	ctx = context.WithValue(ctx, key, astruct{})

	_, err := openmfpctx.GetTenantFromContext(ctx)
	assert.Error(t, err, "someone stored a wrong value in the [tenant] key with type [context.astruct], expected [string]")
}

func TestAddAndGetAuthHeaderToContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		authHeader  any
		expectError bool
	}{
		{
			name:        "should return the auth header from if stored in the context",
			authHeader:  "auth",
			expectError: false,
		},
		{
			name:        "should error out if a wrong value is stored inside the context",
			authHeader:  4,
			expectError: true,
		},
		{
			name:        "should error out if no value is stored inside the context",
			authHeader:  nil,
			expectError: true,
		},
		{
			name:        "should error out if an empty string is stored inside the context",
			authHeader:  "",
			expectError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), keys.AuthHeaderCtxKey, test.authHeader)

			val, err := openmfpctx.GetAuthHeaderFromContext(ctx)
			if test.expectError {
				assert.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, test.authHeader, val)

		})
	}
}

func TestAddWebTokenToContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	issuer := "my-issuer"
	tokenString, err := generateJWT(issuer)
	assert.NoError(t, err)

	ctx = openmfpctx.AddWebTokenToContext(ctx, tokenString)

	token, err := openmfpctx.GetWebTokenFromContext(ctx)
	assert.Nil(t, err)
	assert.Equal(t, issuer, token.Issuer)
}

func TestAddWebTokenToContextNegative(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = context.WithValue(ctx, keys.WebTokenCtxKey, nil)

	_, err := openmfpctx.GetWebTokenFromContext(ctx)
	assert.ErrorContains(t, err, "someone stored a wrong value in the [webToken] key with type [<nil>], expected [jwt.WebToken]")
}

func TestAddWebTokenToContextWrongToken(t *testing.T) {
	t.Parallel()

	initialContext := context.Background()
	tokenString := "not-a-token"

	ctx := openmfpctx.AddWebTokenToContext(initialContext, tokenString)

	assert.Equal(t, initialContext, ctx)
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
	ctx = openmfpctx.AddIsTechnicalIssuerToContext(ctx)

	isTechnicalIssuer := openmfpctx.GetIsTechnicalIssuerFromContext(ctx)
	assert.True(t, isTechnicalIssuer)
}

func TestAddIsTechnicalIssuerToContextNegative(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	isTechnicalIssuer := openmfpctx.GetIsTechnicalIssuerFromContext(ctx)
	assert.False(t, isTechnicalIssuer)
}

func TestHasTenantInContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = openmfpctx.AddTenantToContext(ctx, "tenant")

	hasTenant := openmfpctx.HasTenantInContext(ctx)
	assert.True(t, hasTenant)
}

func TestHasTenantInContextNegative(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	hasTenant := openmfpctx.HasTenantInContext(ctx)
	assert.False(t, hasTenant)
}
