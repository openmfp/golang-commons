package context

import (
	"context"
	"fmt"

	"github.com/openmfp/golang-commons/jwt"
	"github.com/openmfp/golang-commons/logger"
)

type ContextKey string

func AddSpiffeToContext(ctx context.Context, spiffe string) context.Context {
	key := ContextKey(jwt.SpiffeCtxKey)
	return context.WithValue(ctx, key, spiffe)
}

func GetSpiffeFromContext(ctx context.Context) (string, error) {
	key := ContextKey(jwt.SpiffeCtxKey)
	spiffe, ok := ctx.Value(key).(string)
	if !ok {
		return spiffe, fmt.Errorf("someone stored a wrong value in the [%s] key with type [%T], expected [string]", jwt.SpiffeCtxKey, ctx.Value(key))
	}

	return spiffe, nil
}

func AddTenantToContext(ctx context.Context, tenantId string) context.Context {
	key := ContextKey(jwt.TenantIdCtxKey)
	return context.WithValue(ctx, key, tenantId)
}

func GetTenantFromContext(ctx context.Context) (string, error) {
	key := ContextKey(jwt.TenantIdCtxKey)
	tenantId, ok := ctx.Value(key).(string)
	if !ok {
		return tenantId, fmt.Errorf("someone stored a wrong value in the [%s] key with type [%T], expected [string]", jwt.TenantIdCtxKey, ctx.Value(key))
	}

	return tenantId, nil
}

func AddAuthHeaderToContext(ctx context.Context, headerValue string) context.Context {
	key := ContextKey(jwt.AuthHeaderCtxKey)
	return context.WithValue(ctx, key, headerValue)
}

func GetAuthHeaderFromContext(ctx context.Context) (string, error) {
	key := ContextKey(jwt.AuthHeaderCtxKey)
	auth, ok := ctx.Value(key).(string)
	if !ok {
		return auth, fmt.Errorf("someone stored a wrong value in the [%s] key with type [%T], expected [string]", jwt.AuthHeaderCtxKey, ctx.Value(key))
	}

	return auth, nil
}

func AddWebTokenToContext(ctx context.Context, idToken string) context.Context {
	key := ContextKey(jwt.WebTokenCtxKey)
	token, err := jwt.New(idToken)
	if err != nil {
		logger.StdLogger.Error().Err(err).Msg("cannot add given id_token to context")
		return ctx
	}
	return context.WithValue(ctx, key, token)
}

func GetWebTokenFromContext(ctx context.Context) (jwt.WebToken, error) {
	key := ContextKey(jwt.WebTokenCtxKey)
	idToken, ok := ctx.Value(key).(jwt.WebToken)
	if !ok {
		return idToken, fmt.Errorf("someone stored a wrong value in the [%s] key with type [%T], expected [jwt.WebToken]", jwt.WebTokenCtxKey, ctx.Value(key))
	}

	return idToken, nil
}

type isTechnicalUserKey struct{}

func AddIsTechnicalIssuerToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, isTechnicalUserKey{}, true)
}

func GetIsTechnicalIssuerFromContext(ctx context.Context) bool {
	isTechnicalIsser, ok := ctx.Value(isTechnicalUserKey{}).(bool)
	if !ok {
		return false
	}

	return isTechnicalIsser
}
