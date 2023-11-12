package gateway

import (
	"github.com/Bendomey/nucleo-go"
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

	//authorization turn on/off authorization
	Authorization bool

	//authentication turn on/off authentication
	Authentication bool
}

var defaultRoutes = []Route{
	{
		Name: "base-routes",
		Path: "/",

		//whitelist filter used to filter the list of actions.
		//accept regex, and wildcard on action name
		//regex: /^math\.\w+$/
		//wildcard: posts.*
		Whitelist: []string{"**"},

		//mappingPolicy -> all : include all actions, the ones with aliases and without.
		//mappingPolicy -> restrict : include only actions that are in the list of aliases.
		MappingPolicy: MappingPolicyAll,

		//aliases -> alias names instead of action names.
		Aliases: map[string]string{
			// 	"login": "auth.login"
		},

		//authorization turn on/off authorization
		Authorization: false,

		//authentication turn on/off authentication
		Authentication: false,
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
	// "onError": func(context *gin.Context, response nucleo.Payload) {},
}
