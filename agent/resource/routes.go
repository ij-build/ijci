package resource

import (
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

	register("/builds/{build_id:<id>}/cancel", &BuildCancelResource{})
	register("/builds/{build_id:<id>}/logs/{build_log_id:<id>}", &BuildLogResource{})
	return nil
}

func expandTemplate(template string) string {
	return strings.Replace(template, "<id>", consts.PatternUUID, -1)
}
