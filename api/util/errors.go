package util

import (
	"net/http"

	"github.com/efritz/response"
	"github.com/go-nacelle/nacelle"
)

func InternalError(logger nacelle.Logger, err error) response.Response {
	logger.Error(
		"Internal error (%s)",
		err.Error(),
	)

	return response.Empty(http.StatusInternalServerError)
}
