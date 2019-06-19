package util

import (
	"net/http"

	"github.com/go-nacelle/nacelle"
	"github.com/efritz/response"
)

func InternalError(logger nacelle.Logger, err error) response.Response {
	logger.Error(
		"Internal error (%s)",
		err.Error(),
	)

	return response.Empty(http.StatusInternalServerError)
}
