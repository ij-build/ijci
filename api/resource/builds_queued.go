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

type QueuedBuildsResource struct {
	*chevron.EmptySpec
	DB *db.LoggingDB `service:"db"`
}

func (r *QueuedBuildsResource) Get(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	pageMeta, resp := util.GetPageMeta(req)
	if resp != nil {
		return resp
	}

	builds, pagedResultsMeta, err := db.GetQueuedBuilds(
		r.DB,
		pageMeta,
		req.URL.Query().Get("filter"),
	)

	if err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to fetch queued build records (%s)", err.Error()),
		)
	}

	return response.JSON(map[string]interface{}{
		"builds": builds,
		"meta":   pagedResultsMeta,
	})
}
