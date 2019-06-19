package resource

import (
	"context"
	"net/http"

	"github.com/efritz/response"
	"github.com/go-nacelle/chevron"
	"github.com/go-nacelle/nacelle"

	agentctx "github.com/ij-build/ijci/agent/context"
	"github.com/ij-build/ijci/agent/util"
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
