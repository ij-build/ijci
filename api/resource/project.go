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

type ProjectResource struct {
	*chevron.EmptySpec
	DB *db.LoggingDB `service:"db"`
}

func (r *ProjectResource) Get(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	project, err := db.GetProject(r.DB, util.GetProjectID(req))
	if err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to fetch project record (%s)", err.Error()),
		)
	}

	return response.JSON(map[string]interface{}{
		"project": project,
	})
}
