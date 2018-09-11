package resource

import (
	"fmt"

	"github.com/efritz/chevron"
	"github.com/efritz/chevron/middleware"
	"github.com/efritz/nacelle"
)

const uuidPattern = "[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}"

func SetupRoutes(config nacelle.Config, router chevron.Router) error {
	router.AddMiddleware(middleware.NewLogging())
	router.AddMiddleware(middleware.NewRequestID())

	router.MustRegister(
		"/builds",
		&BuildsResource{},
		chevron.WithMiddlewareFor(
			middleware.NewSchemaMiddleware("/schemas/build-post.yaml"),
			chevron.MethodPost,
		),
	)

	router.MustRegister(
		fmt.Sprintf("/builds/{build_id:%s}", uuidPattern),
		&BuildResource{},
		chevron.WithMiddlewareFor(
			middleware.NewSchemaMiddleware("/schemas/build-patch.yaml"),
			chevron.MethodPatch,
		),
	)

	router.MustRegister(
		fmt.Sprintf("/builds/{build_id:%s}/logs", uuidPattern),
		&BuildLogsResource{},
		chevron.WithMiddlewareFor(
			middleware.NewSchemaMiddleware("/schemas/build-log-post.yaml"),
			chevron.MethodPost,
		),
	)

	return nil
}
