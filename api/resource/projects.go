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

type ProjectsResource struct {
	*chevron.EmptySpec
	DB *db.LoggingDB `service:"db"`
}

func (r *ProjectsResource) Get(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	projects, err := db.GetProjects(r.DB)
	if err != nil {
		return util.InternalError(
			logger,
			fmt.Errorf("failed to fetch project records (%s)", err.Error()),
		)
	}

	return response.JSON(map[string]interface{}{
		"projects": projects,
	})
}
