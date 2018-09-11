package resource

import (
	"net/http"

	"github.com/efritz/nacelle"
	"github.com/efritz/response"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func getBuildID(req *http.Request) uuid.UUID {
	return uuid.Must(uuid.Parse(mux.Vars(req)["build_id"]))
}

func internalError(logger nacelle.Logger, err error) response.Response {
	logger.Error(
		"Internal error (%s)",
		err.Error(),
	)

	return response.Empty(http.StatusInternalServerError)
}
