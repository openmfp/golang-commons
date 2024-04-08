package sentry

import (
	"context"
	"errors"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/openmfp/golang-commons/logger"
	"github.com/stretchr/testify/assert"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"

	openmfpcontext "github.com/openmfp/golang-commons/context"
	"github.com/openmfp/golang-commons/jwt"
	testlogger "github.com/openmfp/golang-commons/logger/testlogger"
)

func TestGraphQLRecover(t *testing.T) {
	// Given
	log := testlogger.New()
	recoverFunc := GraphQLRecover(log.Logger)
	ctx := context.WithValue(context.Background(), openmfpcontext.ContextKey(jwt.TenantIdCtxKey), "test")
	ctx = graphql.WithOperationContext(ctx, &graphql.OperationContext{
		Operation: &ast.OperationDefinition{
			Name:      "test",
			Operation: ast.Query,
		},
	})
	ctx = graphql.WithPathContext(ctx, &graphql.PathContext{
		ParentField: &graphql.FieldContext{
			Field: graphql.CollectedField{
				Field: &ast.Field{
					Alias: "test",
					Name:  "test",
				},
			},
		},
	})

	// When
	err := recoverFunc(ctx, "test error")

	// Then
	assert.Equal(t, gqlerror.Errorf("internal server error: test error"), err)
}

func TestGraphQLErrorPresenter(t *testing.T) {
	//Given
	presenter := GraphQLErrorPresenter()
	testError := errors.New("test error")
	ctx := openmfpcontext.AddTenantToContext(context.Background(), "test")

	//When
	err := presenter(ctx, testError)

	//Then
	expectedErr := gqlerror.Wrap(testError)
	assert.Equal(t, expectedErr, err)
}

func TestGraphQLErrorPresenterNilError(t *testing.T) {
	//Given
	presenter := GraphQLErrorPresenter()
	var testError error
	ctx := context.Background()

	//When
	err := presenter(ctx, testError)

	//Then
	assert.Nil(t, err)
}

func TestGraphQLErrorPresenterWithoutTenantContext(t *testing.T) {
	presenter := GraphQLErrorPresenter()
	testError := SentryError(errors.New("test error"))
	ctx := context.Background()

	//When
	err := presenter(ctx, testError)

	//Then
	expectedErr := gqlerror.Wrap(testError)
	assert.Equal(t, expectedErr, err)
}

func TestGraphQLErrorPresenterWithSkipTenants(t *testing.T) {
	//Given
	presenter := GraphQLErrorPresenter("test")
	testError := SentryError(errors.New("test error"))
	tl := testlogger.New()
	ctx := openmfpcontext.AddTenantToContext(context.Background(), "test")
	ctx = logger.SetLoggerInContext(ctx, tl.Logger)

	//When
	err := presenter(ctx, testError)

	//Then
	expectedErr := gqlerror.Wrap(testError)
	assert.Equal(t, expectedErr, err)

	messages, err2 := tl.GetLogMessages()
	assert.NoError(t, err2)
	assert.Len(t, messages, 1)
	assert.Equal(t, "Error not sent to Sentry for skipped tenant", messages[0].Message)
}
