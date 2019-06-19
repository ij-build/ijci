package resource

import (
	"context"
	"fmt"
	"net/http"

	"github.com/efritz/response"
	"github.com/go-nacelle/chevron"
	"github.com/go-nacelle/nacelle"
	"github.com/go-nacelle/pgutil"
	"github.com/ij-build/ijci/api/db"
	"github.com/ij-build/ijci/api/util"
)

type ActiveBuildsResource struct {
	*chevron.EmptySpec
	DB *pgutil.LoggingDB `service:"db"`
}

func (r *ActiveBuildsResource) Get(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	pageMeta, resp := util.GetPageMeta(req)
	if resp != nil {
		return resp
	}

	builds, pagedResultsMeta, err := db.GetActiveBuilds(
		r.DB,
		pageMeta,
		req.URL.Query().Get("filter"),
	)

	if err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to fetch active build records (%s)", err.Error()),
		)
	}

	return response.JSON(map[string]interface{}{
		"builds": builds,
		"meta":   pagedResultsMeta,
	})
}
