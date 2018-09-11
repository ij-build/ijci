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
		fmt.Sprintf("/builds/{build_id:%s}/logs/{build_log_id:%s}", consts.PatternUUID, consts.PatternUUID),
		&BuildLogResource{},
	)

	return nil
}
