package resource

import (
	"fmt"
	"strings"

	"github.com/go-nacelle/chevron"
	"github.com/go-nacelle/chevron/middleware"
	"github.com/go-nacelle/nacelle"
	"github.com/ij-build/ijci/consts"
)

var SetupRoutesFunc = chevron.RouteInitializerFunc(SetupRoutes)

func SetupRoutes(config nacelle.Config, router chevron.Router) error {
	router.AddMiddleware(middleware.NewLogging())
	router.AddMiddleware(middleware.NewRequestID())

	register := func(template string, resource chevron.ResourceSpec, middleware ...chevron.MiddlewareConfigFunc) {
		router.MustRegister(expandTemplate(template), resource, middleware...)
	}

	register("/projects", &ProjectsResource{}, makePostSchema("project"))
	register("/projects/{project_id:<id>}", &ProjectResource{}, makePatchSchema("project"))
	register("/projects/{project_id:<id>}/builds", &ProjectBuildsResource{})
	register("/builds", &BuildsResource{}, makePostSchema("build"))
	register("/builds/active", &ActiveBuildsResource{})
	register("/builds/queued", &QueuedBuildsResource{})
	register("/builds/{build_id:<id>}", &BuildResource{}, makePatchSchema("build"))
	register("/builds/{build_id:<id>}/cancel", &BuildCancelResource{})
	register("/builds/{build_id:<id>}/requeue", &BuildRequeueResource{})
	register("/builds/{build_id:<id>}/logs", &BuildLogsResource{}, makePostSchema("build-log"))
	register("/builds/{build_id:<id>}/logs/{build_log_id:<id>}", &BuildLogResource{}, makePatchSchema("build-log"))
	return nil
}

func expandTemplate(template string) string {
	return strings.Replace(template, "<id>", consts.PatternUUID, -1)
}

func makePostSchema(name string) chevron.MiddlewareConfigFunc {
	return chevron.WithMiddlewareFor(makeSchema(name, "post"), chevron.MethodPost)
}

func makePatchSchema(name string) chevron.MiddlewareConfigFunc {
	return chevron.WithMiddlewareFor(makeSchema(name, "patch"), chevron.MethodPatch)
}

func makeSchema(name, suffix string) chevron.Middleware {
	return middleware.NewSchemaMiddleware(fmt.Sprintf("/schemas/%s-%s.yaml", name, suffix))
}
