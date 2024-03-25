package sentry

import (
	"context"
	"errors"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"

	openmfpcontext "github.com/openmfp/golang-commons/context"
	"github.com/openmfp/golang-commons/jwt"
	"github.com/openmfp/golang-commons/logger"
)

func TestGraphQLRecover(t *testing.T) {
	// Given
	log, _ := logger.NewTestLogger()
	recoverFunc := GraphQLRecover(log)
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
	ctx := context.WithValue(context.Background(), openmfpcontext.ContextKey(jwt.TenantIdCtxKey), "test")

	//When
	err := presenter(ctx, testError)

	//Then
	expectedErr := gqlerror.Wrap(testError)
	assert.Equal(t, expectedErr, err)
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
	ctx := context.WithValue(context.Background(), openmfpcontext.ContextKey(jwt.TenantIdCtxKey), "test")

	//When
	err := presenter(ctx, testError)

	//Then
	expectedErr := gqlerror.Wrap(testError)
	assert.Equal(t, expectedErr, err)

}