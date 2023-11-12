package gateway

import (
	"github.com/Bendomey/nucleo-go"
	"github.com/gin-gonic/gin"
)

type MappingPolicyType string

const (
	LIBRARY_NAME       = "nucleo-gateway"
	LIBRARY_VERSION    = "0.1.0"
	LIBRARY_REPOSITORY = "https://github.com/Bendomey/awesome-gateway"

	MappingPolicyAll      MappingPolicyType = "all"
	MappingPolicyRestrict MappingPolicyType = "restrict"
)

type Route struct {
	// Name of the route group
	Name string

	// Path of the route group
	Path string

	// Route-level middlewares.
	Use []gin.HandlerFunc

	//whitelist filter used to filter the list of actions.
	//accept regex, and wildcard on action name
	//regex: /^math\.\w+$/
	//wildcard: posts.*
	Whitelist []string

	//mappingPolicy -> all : include all actions, the ones with aliases and without.
	//mappingPolicy -> restrict : include only actions that are in the list of aliases.
	MappingPolicy MappingPolicyType

	//aliases -> alias names instead of action names.
	Aliases map[string]string

	// This is called before action is called
	OnBeforeCall *func(context nucleo.Context, ginContext *gin.Context, route Route, alias string)

	// This is called after action is called but before response is sent to user.
	OnAfterCall *func(context nucleo.Context, ginContext *gin.Context, route Route, response nucleo.Payload)

	//authorization turn on/off authorization
	Authorization bool

	//authentication turn on/off authentication
	Authentication bool
}

var defaultRoutes = []Route{
	{
		Name: "base-routes",
		Path: "/",

		Use: []gin.HandlerFunc{},

		Whitelist: []string{"**"},

		MappingPolicy: MappingPolicyAll,

		Aliases: map[string]string{},

		OnBeforeCall: nil,

		OnAfterCall: nil,
	},
}

// Default settings for the gateway service
var defaultSettings = map[string]interface{}{
	// Exposed port
	"port": 5000,

	// Exposed IP
	"ip": "0.0.0.0",

	// base path
	"path": "/",

	// Global middlewares. Applied to all routes.
	"use": []gin.HandlerFunc{},

	// Routes
	"routes": defaultRoutes,

	// Log each request (default to "info" level)
	"logRequest": nucleo.LogLevelDebug,

	// Log the request ctx.params (default to "debug" level)
	"logRequestParams": nucleo.LogLevelInfo,

	// Log each response (default to "info" level)
	"logResponse": nucleo.LogLevelInfo,

	// Log the response data (default to disable)
	"logResponseData": nucleo.LogLevelInfo,

	// If set to true, it will log 4xx client errors, as well
	"log4XXResponses": nucleo.LogLevelInfo,

	// Log the route registration/aliases related activity
	"logRouteRegistration": nucleo.LogLevelInfo,

	// Optimize route order
	"optimizeOrder": true,

	// FIXME: parsing issue in nucleo-go
	"onError": func(context *gin.Context, response nucleo.Payload) {},
}
