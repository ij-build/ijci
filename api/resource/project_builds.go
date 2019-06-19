package resource

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-nacelle/chevron"
	"github.com/go-nacelle/nacelle"
	"github.com/efritz/response"

	"github.com/ij-build/ijci/api/db"
	"github.com/ij-build/ijci/api/util"
)

type ProjectBuildsResource struct {
	*chevron.EmptySpec
	DB *db.LoggingDB `service:"db"`
}

func (r *ProjectBuildsResource) Get(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	pageMeta, resp := util.GetPageMeta(req)
	if resp != nil {
		return resp
	}

	builds, pagedResultsMeta, err := db.GetBuildsForProject(
		r.DB,
		util.GetProjectID(req),
		pageMeta,
		req.URL.Query().Get("filter"),
	)

	if err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to fetch build records (%s)", err.Error()),
		)
	}

	return response.JSON(map[string]interface{}{
		"builds": builds,
		"meta":   pagedResultsMeta,
	})
}
