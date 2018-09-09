package http

import (
	"net/http"

	"github.com/efritz/nacelle"
	"github.com/efritz/response"
)

func internalError(logger nacelle.Logger, err error) response.Response {
	logger.Error(
		"Internal error (%s)",
		err.Error(),
	)

	return response.Empty(http.StatusInternalServerError)
}
