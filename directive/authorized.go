package directive

import (
	"context"
	"encoding/json"
	"fmt"
	openmfperrors "github.com/openmfp/golang-commons/errors"
	"github.com/openmfp/golang-commons/sentry"
	"github.com/rs/zerolog/log"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	openmfpcontext "github.com/openmfp/golang-commons/context"
	"github.com/openmfp/golang-commons/fga/helpers"
	"github.com/openmfp/golang-commons/logger"
	"google.golang.org/grpc/metadata"
)

func extractNestedKeyFromArgs(args map[string]any, paramName string) (string, error) {
	o, err := json.Marshal(args)
	if err != nil {
		return "", err
	}

	var normalizedArgs map[string]any
	err = json.Unmarshal(o, &normalizedArgs)
	if err != nil {
		return "", err
	}

	var paramValue string
	parts := strings.Split(paramName, ".")
	for i, key := range parts {
		val, ok := normalizedArgs[key]
		if !ok {
			return "", fmt.Errorf("unable to extract param from request for given paramName %q", paramName)
		}

		if i == len(strings.Split(paramName, "."))-1 {
			paramValue, ok = val.(string)
			if !ok || paramValue == "" {
				return "", fmt.Errorf("unable to extract param from request for given paramName %q, param is of wrong type", paramName)
			}

			return paramValue, nil
		}

		normalizedArgs, ok = val.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("unable to extract param from request for given paramName %q, param is of wrong type", paramName)
		}
	}

	return paramValue, nil
}

func Authorized(openfgaClient openfgav1.OpenFGAServiceClient, log *logger.Logger) func(context.Context, interface{}, graphql.Resolver, string, *string, *string, string) (interface{}, error) {
	compLogger := log.ComponentLogger("authorizedDirective")
	ac := authChecker{
		log:           compLogger,
		openfgaClient: openfgaClient,
	}

	if !directiveConfiguration.DirectivesAuthorizationEnabled {
		log.Trace().Msg("Authorization directive is disabled. Skipping authorization check.")
		return func(ctx context.Context, obj interface{}, next graphql.Resolver, relation string, entityType *string, entityTypeParamName *string, entityParamName string) (interface{}, error) {
			return next(ctx)
		}
	}

	return func(ctx context.Context, obj interface{}, next graphql.Resolver, relation string, entityType *string, entityTypeParamName *string, entityParamName string) (interface{}, error) {
		if openfgaClient == nil {
			return nil, sentry.SentryError(openmfperrors.New("OpenFGAServiceClient is nil. Cannot process request"))
		}

		ctx, hasToken, err := ac.withTenantContextForTechnicalUsers(ctx)
		if err != nil {
			compLogger.Info().Err(err).Msg("error setting tenant context for technical users")
			return nil, err
		}

		entityID, tenantID, evaluatedEntityType, err := ac.prepareAuthCheckInputs(ctx, entityParamName, entityTypeParamName, entityType)
		if err != nil {
			compLogger.Info().Err(err).Msg("error when extracting values for auth check")
			return nil, err
		}

		res, err := ac.executeTheAuthCheck(ctx, hasToken, entityID, tenantID, evaluatedEntityType, relation)
		if err != nil {
			compLogger.Error().Err(err).Msg("error in authorized directive")
			return nil, sentry.SentryError(err)
		}

		if !res.Allowed {
			log.Info().Bool("allowed", res.Allowed).Msg("not allowed")
			return nil, gqlerror.Errorf("unauthorized")
		}

		return next(ctx)
	}
}

type authChecker struct {
	log           *logger.Logger
	openfgaClient openfgav1.OpenFGAServiceClient
}

func (ac *authChecker) executeTheAuthCheck(ctx context.Context, hasToken bool, entityID string, tenantID string, evaluatedEntityType string, relation string) (*openfgav1.CheckResponse, error) {
	storeID, err := helpers.GetStoreIDForTenant(ctx, ac.openfgaClient, tenantID)
	if err != nil {
		return nil, err
	}
	modelID, err := helpers.GetModelIDForTenant(ctx, ac.openfgaClient, tenantID)
	if err != nil {
		return nil, err
	}

	var userID string
	if hasToken {
		user, err := openmfpcontext.GetWebTokenFromContext(ctx)
		if err != nil {
			return nil, err
		}
		userID = user.Subject
	} else {
		spiffe, err := openmfpcontext.GetSpiffeFromContext(ctx)
		if err != nil {
			return nil, openmfperrors.New("authorized was invoked without a user token or a spiffe header")
		}
		userID = strings.TrimPrefix(spiffe, "spiffe://")
		log.Trace().Str("user", userID).Msg("using spiffe user in authorized directive")
	}

	req := &openfgav1.CheckRequest{
		StoreId:              storeID,
		AuthorizationModelId: modelID,
		TupleKey: &openfgav1.CheckRequestTupleKey{
			User:     fmt.Sprintf("user:%s", helpers.SanitizeUserID(userID)),
			Relation: relation,
			Object:   fmt.Sprintf("%s:%s", evaluatedEntityType, entityID),
		},
	}

	res, err := ac.openfgaClient.Check(ctx, req)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, openmfperrors.New("received nil response from openfgaClient.Check with no error")
	}
	return res, nil
}

func (ac *authChecker) withTenantContextForTechnicalUsers(ctx context.Context) (context.Context, bool, error) {
	newCtx, err := setTenantToContextForTechnicalUsers(ctx, ac.log)
	if err != nil {
		return ctx, false, openmfperrors.EnsureStack(err)
	}

	token, err := openmfpcontext.GetAuthHeaderFromContext(newCtx)
	hasToken := err == nil

	if hasToken {
		newCtx = metadata.AppendToOutgoingContext(newCtx, "authorization", token)
	}

	return newCtx, hasToken, nil
}

func (ac *authChecker) prepareAuthCheckInputs(
	ctx context.Context,
	entityParamName string,
	entityTypeParamName *string,
	entityType *string,
) (
	entityID string,
	tenantID string,
	evaluatedEntityType string,
	err error,
) {
	fctx := graphql.GetFieldContext(ctx)

	entityID, err = extractNestedKeyFromArgs(fctx.Args, entityParamName)
	if err != nil {
		err = openmfperrors.EnsureStack(err)
		return
	}

	tenantID, err = openmfpcontext.GetTenantFromContext(ctx)
	if err != nil {
		err = openmfperrors.EnsureStack(err)
		return
	}

	if entityTypeParamName != nil {
		evaluatedEntityType, err = extractNestedKeyFromArgs(fctx.Args, *entityTypeParamName)
		if err != nil {
			err = openmfperrors.EnsureStack(err)
			return
		}
	} else if entityType != nil {
		evaluatedEntityType = *entityType
	}

	if evaluatedEntityType == "" {
		err = openmfperrors.New("make sure to either provide entityType or entityTypeParamName")
		return
	}
	return
}
