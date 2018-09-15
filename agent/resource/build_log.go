package resource

import (
	"context"
	"fmt"
	"net/http"

	"github.com/efritz/chevron"
	"github.com/efritz/nacelle"
	"github.com/efritz/response"

	"github.com/efritz/ijci/agent/logs"
	"github.com/efritz/ijci/util"
)

type BuildLogResource struct {
	*chevron.EmptySpec
	LogProcessor *logs.LogProcessor `service:"log-processor"`
}

func (r *BuildLogResource) Get(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	streamCtx, cancel := context.WithCancel(context.Background())

	ch, err := r.LogProcessor.GetBuildLogStream(
		streamCtx,
		util.GetBuildID(req),
		util.GetBuildLogID(req),
	)

	if err != nil {
		cancel()

		if err == logs.ErrUnknownBuildLog {
			return response.Empty(http.StatusNotFound)
		}

		return util.InternalError(
			logger,
			fmt.Errorf("failed to get build log stream (%s)", err.Error()),
		)
	}

	return response.Stream(util.NewChanReader(ch), response.WithFlush()).AddCallback(func(_ error) {
		cancel()
	})
}
