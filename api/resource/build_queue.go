package resource

import (
	"context"
	"fmt"
	"net/http"

	"github.com/efritz/chevron"
	"github.com/efritz/nacelle"
	"github.com/efritz/response"

	"github.com/efritz/ijci/api/db"
	"github.com/efritz/ijci/util"
)

type BuildQueueResource struct {
	*chevron.EmptySpec
	DB *db.LoggingDB `service:"db"`
}

func (r *BuildQueueResource) Get(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	queuedBuilds, err := db.GetQueuedBuilds(r.DB)
	if err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to fetch queued build records (%s)", err.Error()),
		)
	}

	activeBuilds, err := db.GetActiveBuilds(r.DB)
	if err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to fetch active build records (%s)", err.Error()),
		)
	}

	return response.JSON(map[string]interface{}{
		"queued": queuedBuilds,
		"active": activeBuilds,
	})
}
