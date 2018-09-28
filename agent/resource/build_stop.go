package resource

import (
	"context"
	"net/http"

	"github.com/efritz/chevron"
	"github.com/efritz/nacelle"
	"github.com/efritz/response"

	agentctx "github.com/efritz/ijci/agent/context"
	"github.com/efritz/ijci/agent/util"
)

type BuildCancelResource struct {
	*chevron.EmptySpec
	ContextProcessor *agentctx.ContextProcessor `service:"context-processor"`
}

func (r *BuildCancelResource) Post(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	if err := r.ContextProcessor.Cancel(util.GetBuildID(req)); err != nil {
		return response.Empty(http.StatusNotFound)
	}

	return response.Empty(http.StatusOK)
}
