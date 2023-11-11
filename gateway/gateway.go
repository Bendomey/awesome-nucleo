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

type AuthMethodsFunc = func(context nucleo.Context, request *http.Request, route string)

type GatewayService struct {
	Authentication AuthMethodsFunc
	Authorization  AuthMethodsFunc
	settings       map[string]interface{}
	mainRouter     *gin.Engine
	gatewayRouter  *gin.RouterGroup
	server         *http.Server
}

type GatewayMixin struct {
	Authentication AuthMethodsFunc
	Authorization  AuthMethodsFunc
}

func NewGatewayMixin(start GatewayMixin) nucleo.Mixin {
	gatewayMixin := GatewayService{
		Authentication: start.Authentication,
		Authorization:  start.Authorization,
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
	}

	// we have a global path that user's can set their gateways up with.
	svc.registerBaseGatewayPath()

	go svc.startServer(context)
	go registerActionsRouter(context.(nucleo.Context), svc.settings, svc.gatewayRouter)

	context.Logger().Info("Gateway Started()")
}

func (svc *GatewayService) Stopped(context nucleo.BrokerContext, service nucleo.ServiceSchema) {
	if svc.server != nil {
		ctx, cancel := goContext.WithTimeout(goContext.Background(), 5*time.Second)
		defer cancel()

		if err := svc.server.Shutdown(ctx); err != nil {
			context.Logger().Error("Error shutting down server - error: ", err)
		}

	}
}

func (svc *GatewayService) getAddress() string {
	ip := svc.settings["ip"].(string)
	port := svc.settings["port"].(int)
	return fmt.Sprint(ip, ":", port)
}

func (svc *GatewayService) startServer(context nucleo.BrokerContext) {
	address := svc.getAddress()
	context.Logger().Info("Server starting to listen on: ", address)

	if err := svc.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		context.Logger().Error("Error listening server on: ", address, " error: ", err)
		return
	}

	context.Logger().Info("Server started on: ", address)
}

func (svc *GatewayService) registerBaseGatewayPath() {

	basePath, exists := svc.settings["path"].(string)
	path := "/"
	if exists {
		path = basePath
	}
	svc.gatewayRouter = svc.mainRouter.Group(path)
}
