package gateway

import (
	goContext "context"
	"fmt"
	"net/http"
	"time"

	"github.com/Bendomey/nucleo-go"
	"github.com/Bendomey/nucleo-go/service"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type AuthenticateMethodsFunc = func(context nucleo.Context, ginContext *gin.Context, alias string) interface{}
type AuthorizeMethodFunc = func(context nucleo.Context, ginContext *gin.Context, alias string)

type GatewayService struct {
	Authenticate  *AuthenticateMethodsFunc
	Authorize     *AuthorizeMethodFunc
	settings      map[string]interface{}
	mainRouter    *gin.Engine
	gatewayRouter *gin.RouterGroup
	server        *http.Server
}

type GatewayMixin struct {
	Authenticate *AuthenticateMethodsFunc
	Authorize    *AuthorizeMethodFunc
}

func NewGatewayMixin(start GatewayMixin) nucleo.Mixin {
	gatewayMixin := GatewayService{
		Authenticate: start.Authenticate,
		Authorize:    start.Authorize,
	}

	return nucleo.Mixin{
		Name:         gatewayMixin.Name(),
		Dependencies: gatewayMixin.Dependencies(),
		Settings:     gatewayMixin.settings,
		Metadata:     gatewayMixin.Metadata(),
		Created:      gatewayMixin.Created,
		Started:      gatewayMixin.Started,
		Stopped:      gatewayMixin.Stopped,
	}
}

func (svc GatewayService) Name() string {
	return "gateway"
}

func (svc *GatewayService) Dependencies() []string {
	return []string{"$node"}
}

func (svc *GatewayService) Metadata() map[string]interface{} {
	return map[string]interface{}{
		"$category":    "gateway",
		"$description": "Official API Gateway service",
		"$official":    true,
		"$package": map[string]interface{}{
			"name":    LIBRARY_NAME,
			"version": LIBRARY_VERSION,
			"repo":    LIBRARY_REPOSITORY,
		},
	}
}

func (svc *GatewayService) Created(schema nucleo.ServiceSchema, logger *log.Entry) {
	// Merge user defined settings with our default settings
	svc.settings = service.MergeSettings(defaultSettings, schema.Settings, svc.settings)
}

func (svc *GatewayService) Started(context nucleo.BrokerContext, schema nucleo.ServiceSchema) {
	// create gin server
	svc.mainRouter = gin.Default()

	address := svc.getAddress()
	svc.server = &http.Server{
		Addr:    address,
		Handler: svc.mainRouter,
		// TODO: setup Https support
	}

	// register all global middlewares
	svc.registerGlobalMiddlewares()

	// we have a global path that user's can set their gateways up with.
	svc.registerBaseGatewayPath()

	go svc.startServer(context)
	go svc.registerActionsRouter(context.(nucleo.Context))
	context.Logger().Infoln("Gateway Started()")
}

func (svc *GatewayService) Stopped(context nucleo.BrokerContext, service nucleo.ServiceSchema) {
	if svc.server != nil {
		ctx, cancel := goContext.WithTimeout(goContext.Background(), 5*time.Second)
		defer cancel()

		if err := svc.server.Shutdown(ctx); err != nil {
			context.Logger().Errorln("Error shutting down server - error: ", err)
		}

	}
}

// registerActionsRouter registers all exposed permitted actions/aliases as REST endpoints.
func (svc *GatewayService) registerActionsRouter(context nucleo.Context) {
	// make sure we have a router
	if svc.gatewayRouter == nil {
		return
	}

	for _, actionHandler := range svc.getPermittedActionsAndThenCreateEndpoints(context, fetchServices(context)) {
		actionHandler.context = context
		actionHandler.settings = svc.settings

		path := actionHandler.getFullPath()
		context.Logger().Traceln("registerActionsRouter() action -> ", actionHandler.action, " path: ", path)

		methods := actionHandler.AcceptedMethods()

		// loop over methods
		for method, shouldRegisterMethod := range methods {
			if shouldRegisterMethod {
				actionHandler.router.Handle(method, path, actionHandler.Handler())
			}
		}
	}

}

func (svc *GatewayService) getPermittedActionsAndThenCreateEndpoints(context nucleo.Context, services []map[string]interface{}) []*actionHandler {
	actionHandlers := []*actionHandler{}

	// get the list of routes
	routes := svc.settings["routes"].([]Route)

	for _, route := range routes {
		filteredActions := []string{}

		settingsWhiteList := route.Whitelist
		whitelist := []string{"**"}
		if settingsWhiteList != nil {
			whitelist = settingsWhiteList
		}

		for _, service := range services {
			actions := service["actions"].(map[string]map[string]interface{})
			for _, action := range actions {
				actionName := action["name"].(string)
				if shouldIncludeAction(whitelist, actionName) {
					filteredActions = append(filteredActions, actionName)
				}
			}
		}

		middlewares := []gin.HandlerFunc{}
		if route.Use != nil {
			middlewares = route.Use
		}

		settingsRoutePath := route.Path
		routePath := "/"
		if settingsRoutePath != "" {
			routePath = settingsRoutePath
		}

		// create a route
		newRouterGroup := svc.gatewayRouter.Group(routePath)

		//register middlewares
		newRouterGroup.Use(middlewares...)

		// now that we have the permitted actions, we gotta create the REST endpoints
		actionHandlers = append(actionHandlers, createActionHandlers(route, filteredActions, newRouterGroup, svc.Authenticate, svc.Authorize)...)
	}

	return actionHandlers
}

func (svc *GatewayService) getAddress() string {
	ip := svc.settings["ip"].(string)
	port := svc.settings["port"].(int)
	return fmt.Sprint(ip, ":", port)
}

func (svc *GatewayService) startServer(context nucleo.BrokerContext) {
	address := svc.getAddress()
	context.Logger().Infoln("Server starting to listen on: ", address)

	if err := svc.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		context.Logger().Errorln("Error listening server on: ", address, " error: ", err)
		return
	}

	context.Logger().Infoln("Server started on: ", address)
}

func (svc *GatewayService) registerGlobalMiddlewares() {

	middlewaresSettings, exists := svc.settings["use"].([]gin.HandlerFunc)
	middlewares := []gin.HandlerFunc{}
	if exists {
		middlewares = middlewaresSettings
	}
	svc.mainRouter.Use(middlewares...)
}

func (svc *GatewayService) registerBaseGatewayPath() {

	basePath, exists := svc.settings["path"].(string)
	path := "/"
	if exists {
		path = basePath
	}
	svc.gatewayRouter = svc.mainRouter.Group(path)
}
