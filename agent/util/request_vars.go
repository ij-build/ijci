package util

import (
	"net/http"

	"github.com/go-nacelle/nacelle"
	"github.com/efritz/response"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func GetBuildID(req *http.Request) uuid.UUID {
	return uuid.Must(uuid.Parse(mux.Vars(req)["build_id"]))
}

func GetBuildLogID(req *http.Request) uuid.UUID {
	return uuid.Must(uuid.Parse(mux.Vars(req)["build_log_id"]))
}

func InternalError(logger nacelle.Logger, err error) response.Response {
	logger.Error(
		"Internal error (%s)",
		err.Error(),
	)

	return response.Empty(http.StatusInternalServerError)
}
