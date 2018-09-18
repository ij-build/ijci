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
	builds, err := db.GetBuildQueue(r.DB)
	if err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to fetch build records (%s)", err.Error()),
		)
	}

	return response.JSON(map[string]interface{}{
		"builds": builds,
	})
}
