package http

import (
	"github.com/efritz/chevron"
	"github.com/efritz/chevron/middleware"
	"github.com/efritz/nacelle"
)

func SetupRoutes(config nacelle.Config, router chevron.Router) error {
	router.AddMiddleware(middleware.NewLogging())
	router.AddMiddleware(middleware.NewRequestID())

	router.MustRegister(
		"/builds",
		&BuildsResource{},
		chevron.WithMiddlewareFor(
			middleware.NewSchemaMiddleware("/schemas/build-request.yaml"),
			chevron.MethodPost,
		),
	)

	router.MustRegister(
		"/builds/{build_id}",
		&BuildResource{},
		chevron.WithMiddlewareFor(
			middleware.NewSchemaMiddleware("/schemas/build-request.yaml"),
			chevron.MethodPost,
		),
	)

	return nil
}
