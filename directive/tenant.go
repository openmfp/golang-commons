package directive

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	openmfpcontext "github.com/openmfp/golang-commons/context"
	"github.com/openmfp/golang-commons/logger"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func setTenantToContextForTechnicalUsers(ctx context.Context, l *logger.Logger) (context.Context, error) {
	spiffee, err := openmfpcontext.GetSpiffeFromContext(ctx)
	hasSpiffee := err == nil && spiffee != ""
	if isTechnicalIssuer := openmfpcontext.GetIsTechnicalIssuerFromContext(ctx); !isTechnicalIssuer && !hasSpiffee {
		return ctx, nil
	}

	fieldContext := graphql.GetFieldContext(ctx)
	var tenantID string
	switch tID := fieldContext.Args["tenantId"].(type) {
	case string:
		tenantID = tID
	case *string:
		if tID == nil {
			return nil, &gqlerror.Error{Message: "tenantId parameter is nil - bad request"}
		}
		tenantID = *tID
	}

	if tenantID == "" {
		return ctx, nil
	}

	ctx = openmfpcontext.AddTenantToContext(ctx, tenantID)
	l.Debug().Str("tenantId", tenantID).Msg("Added a tenant id for technical user to the context")
	return ctx, nil
}
