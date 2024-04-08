package context

import (
	"context"
	"fmt"

	"github.com/openmfp/golang-commons/context/keys"
	"github.com/openmfp/golang-commons/jwt"
	"github.com/openmfp/golang-commons/logger"
)

type ContextKey string

func AddSpiffeToContext(ctx context.Context, spiffe string) context.Context {
	return context.WithValue(ctx, keys.SpiffeCtxKey, spiffe)
}

func GetSpiffeFromContext(ctx context.Context) (string, error) {
	spiffe, ok := ctx.Value(keys.SpiffeCtxKey).(string)
	if !ok {
		return spiffe, fmt.Errorf("someone stored a wrong value in the [%s] key with type [%T], expected [string]", jwt.SpiffeCtxKey, ctx.Value(keys.SpiffeCtxKey))
	}

	return spiffe, nil
}

func AddTenantToContext(ctx context.Context, tenantId string) context.Context {
	return context.WithValue(ctx, keys.TenantIdCtxKey, tenantId)
}

func GetTenantFromContext(ctx context.Context) (string, error) {
	tenantId, ok := ctx.Value(keys.TenantIdCtxKey).(string)
	if !ok {
		return tenantId, fmt.Errorf("someone stored a wrong value in the [%s] key with type [%T], expected [string]", jwt.TenantIdCtxKey, ctx.Value(keys.TenantIdCtxKey))
	}

	return tenantId, nil
}

func HasTenantInContext(ctx context.Context) bool {
	_, ok := ctx.Value(keys.TenantIdCtxKey).(string)
	return ok
}

func AddAuthHeaderToContext(ctx context.Context, headerValue string) context.Context {
	return context.WithValue(ctx, keys.AuthHeaderCtxKey, headerValue)
}

func GetAuthHeaderFromContext(ctx context.Context) (string, error) {
	auth, ok := ctx.Value(keys.AuthHeaderCtxKey).(string)
	if !ok {
		return auth, fmt.Errorf("someone stored a wrong value in the [%s] key with type [%T], expected [string]", jwt.AuthHeaderCtxKey, ctx.Value(keys.AuthHeaderCtxKey))
	}

	return auth, nil
}

func AddWebTokenToContext(ctx context.Context, idToken string) context.Context {
	token, err := jwt.New(idToken)
	if err != nil {
		logger.StdLogger.Error().Err(err).Msg("cannot add given id_token to context")
		return ctx
	}
	return context.WithValue(ctx, keys.WebTokenCtxKey, token)
}

func GetWebTokenFromContext(ctx context.Context) (jwt.WebToken, error) {
	idToken, ok := ctx.Value(keys.WebTokenCtxKey).(jwt.WebToken)
	if !ok {
		return idToken, fmt.Errorf("someone stored a wrong value in the [%s] key with type [%T], expected [jwt.WebToken]", jwt.WebTokenCtxKey, ctx.Value(keys.WebTokenCtxKey))
	}

	return idToken, nil
}

func AddIsTechnicalIssuerToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, keys.TechnicalUserCtxKey, true)
}

func GetIsTechnicalIssuerFromContext(ctx context.Context) bool {
	isTechnicalIsser, ok := ctx.Value(keys.TechnicalUserCtxKey).(bool)
	if !ok {
		return false
	}

	return isTechnicalIsser
}
