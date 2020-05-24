package resource

import (
	"context"
	"net/http"

	"github.com/efritz/response"
	"github.com/efritz/sse"
	"github.com/go-nacelle/chevron"
	"github.com/go-nacelle/nacelle"
	"github.com/ij-build/ijci/api/db"
)

type SSEResource struct {
	*chevron.EmptySpec
	sseServer *sse.Server
	Monitor   db.Monitor `service:"monitor"`
}

func (r *SSEResource) PostInject() error {
	ch := make(chan interface{})

	go func() {
		id, events := r.Monitor.Subscribe()
		defer r.Monitor.Unsubscribe(id)

		for event := range events {
			// TODO - map these values (need to fetch build logs after update - try to do this upstream if possible)
			// TODO - add filter in sse package
			ch <- event
		}
	}()

	r.sseServer = sse.NewServer(ch)
	go r.sseServer.Start()
	return nil
}

func (r *SSEResource) Get(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	return r.sseServer.Handler(req)
}
