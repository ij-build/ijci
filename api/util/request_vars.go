package util

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func GetProjectID(req *http.Request) uuid.UUID {
	return uuid.Must(uuid.Parse(mux.Vars(req)["project_id"]))
}

func GetBuildID(req *http.Request) uuid.UUID {
	return uuid.Must(uuid.Parse(mux.Vars(req)["build_id"]))
}

func GetBuildLogID(req *http.Request) uuid.UUID {
	return uuid.Must(uuid.Parse(mux.Vars(req)["build_log_id"]))
}
