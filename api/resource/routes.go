package resource

import (
	"fmt"

	"github.com/efritz/chevron"
	"github.com/efritz/chevron/middleware"
	"github.com/efritz/nacelle"

	"github.com/efritz/ijci/consts"
)

var SetupRoutesFunc = chevron.RouteInitializerFunc(SetupRoutes)

func SetupRoutes(config nacelle.Config, router chevron.Router) error {
	router.AddMiddleware(middleware.NewLogging())
	router.AddMiddleware(middleware.NewRequestID())

	router.MustRegister(
		"/projects",
		&ProjectsResource{},
		chevron.WithMiddlewareFor(
			middleware.NewSchemaMiddleware("/schemas/project-post.yaml"),
			chevron.MethodPost,
		),
	)

	router.MustRegister(
		fmt.Sprintf("/projects/{project_id:%s}", consts.PatternUUID),
		&ProjectResource{},
		chevron.WithMiddlewareFor(
			middleware.NewSchemaMiddleware("/schemas/project-patch.yaml"),
			chevron.MethodPatch,
		),
	)

	router.MustRegister(
		"/builds",
		&BuildsResource{},
		chevron.WithMiddlewareFor(
			middleware.NewSchemaMiddleware("/schemas/build-post.yaml"),
			chevron.MethodPost,
		),
	)

	router.MustRegister(
		fmt.Sprintf("/builds/{build_id:%s}", consts.PatternUUID),
		&BuildResource{},
		chevron.WithMiddlewareFor(
			middleware.NewSchemaMiddleware("/schemas/build-patch.yaml"),
			chevron.MethodPatch,
		),
	)

	router.MustRegister(
		fmt.Sprintf("/builds/{build_id:%s}/requeue", consts.PatternUUID),
		&BuildRequeueResource{},
	)

	router.MustRegister(
		fmt.Sprintf("/builds/{build_id:%s}/logs", consts.PatternUUID),
		&BuildLogsResource{},
		chevron.WithMiddlewareFor(
			middleware.NewSchemaMiddleware("/schemas/build-log-post.yaml"),
			chevron.MethodPost,
		),
	)

	router.MustRegister(
		fmt.Sprintf("/builds/{build_id:%s}/logs/{build_log_id:%s}", consts.PatternUUID, consts.PatternUUID),
		&BuildLogResource{},
		chevron.WithMiddlewareFor(
			middleware.NewSchemaMiddleware("/schemas/build-log-patch.yaml"),
			chevron.MethodPatch,
		),
	)

	return nil
}
