package endpoints

import (
	"github.com/efritz/chevron"
	"github.com/efritz/chevron/middleware"
	"github.com/efritz/nacelle"
)

func SetupRoutes(config nacelle.Config, router chevron.Router) error {
	router.AddMiddleware(middleware.NewLogging())
	router.AddMiddleware(middleware.NewRequestID())

	router.MustRegister(
		"/hook",
		&HookResource{},
		chevron.WithMiddlewareFor(
			middleware.NewSchemaMiddleware("/schemas/repo.yaml"),
			chevron.MethodPost,
		),
	)

	return nil
}
