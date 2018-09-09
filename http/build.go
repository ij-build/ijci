package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/efritz/chevron"
	"github.com/efritz/nacelle"
	"github.com/efritz/response"
	"github.com/gorilla/mux"
)

type BuildResource struct {
	*chevron.EmptySpec

	Logger nacelle.Logger `service:"logger"`
}

func (br *BuildResource) Get(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	// TODO - get record from db
	fmt.Printf(">> %s\n", mux.Vars(req)["build_id"])

	return response.Empty(http.StatusOK)
}
