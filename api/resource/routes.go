package resource

import (
	"fmt"
	"strings"

	"github.com/efritz/chevron"
	"github.com/efritz/chevron/middleware"
	"github.com/efritz/nacelle"

	"github.com/efritz/ijci/consts"
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
	register("/builds", &BuildsResource{}, makePostSchema("build"))
	register("/builds/{build_id:<id>}", &BuildResource{}, makePatchSchema("build"))
	register("/builds/{build_id:<id>}/stop", &BuildStopResource{})
	register("/builds/{build_id:<id>}/requeue", &BuildRequeueResource{})
	register("/builds/{build_id:<id>}/logs", &BuildLogsResource{}, makePostSchema("build-log"))
	register("/builds/{build_id:<id>}/logs/{build_log_id:<id>}", &BuildLogResource{}, makePatchSchema("build-log"))
	register("/queue", &BuildQueueResource{})
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
