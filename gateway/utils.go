package gateway

import (
	"regexp"

	"github.com/Bendomey/nucleo-go"
	"github.com/gin-gonic/gin"
)

// registerActionsRouter registers all exposed permitted actions/aliases as REST endpoints.
func registerActionsRouter(context nucleo.Context, settings map[string]interface{}, router *gin.RouterGroup) {
	// make sure we have a router
	if router == nil {
		return
	}

	for _, actionHandler := range getPermittedActionsAndThenCreateEndpoints(context, settings, fetchServices(context)) {
		actionHandler.context = context

		path := actionHandler.getFullPath()
		context.Logger().Trace("registerActionsRouter() action -> ", actionHandler.action, " path: ", path)

		methods := actionHandler.AcceptedMethods()

		// loop over methods
		for method, shouldRegisterMethod := range methods {
			if shouldRegisterMethod {
				router.Handle(method, path, actionHandler.Handler())
			}
		}
	}

}

// fetch available services with their actions from nucleo registry. Yeah, it's nice like that ;)
func fetchServices(context nucleo.Context) []map[string]interface{} {
	services := <-context.Call("$node.services", map[string]interface{}{
		"onlyAvailable": true,
		"withActions":   true,
	})

	if services.IsError() {
		context.Logger().Error("Could not load the list of services/action from the registry. Error: ", services.Error())
		return []map[string]interface{}{}
	}

	return services.MapArray()
}

func getPermittedActionsAndThenCreateEndpoints(context nucleo.Context, settings map[string]interface{}, services []map[string]interface{}) []*actionHandler {
	actionHandlers := []*actionHandler{}

	// get the list of routes
	routes := settings["routes"].([]Route)

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

		// now that we have the permitted actions, we gotta create the REST endpoints
		actionHandlers = append(actionHandlers, createActionHandlers(route, filteredActions)...)
	}

	return actionHandlers
}

var actionWildCardRegex = regexp.MustCompile(`(.+)\.\*`)
var serviceWildCardRegex = regexp.MustCompile(`\*\.(.+)`)
var serviceActionRegex = regexp.MustCompile(`(.+)\.(.+)`)

// shouldInclude check if the actions should be added based on the whitelist.
func shouldIncludeAction(whitelist []string, action string) bool {
	for _, item := range whitelist {
		if item == "**" || item == "*.*" {
			return true
		}
		whitelistService := actionWildCardRegex.FindStringSubmatch(item)
		if len(whitelistService) > 0 && whitelistService[1] != "" {
			actionService := serviceActionRegex.FindStringSubmatch(action)
			if len(actionService) > 1 && len(whitelistService) > 1 && actionService[1] == whitelistService[1] {
				return true
			}
		}
		whitelistAction := serviceWildCardRegex.FindStringSubmatch(item)
		if len(whitelistAction) > 0 && whitelistAction[1] != "" {
			actionName := serviceActionRegex.FindStringSubmatch(action)
			if len(actionName) > 2 && len(whitelistAction) > 1 && actionName[2] == whitelistAction[1] {
				return true
			}
		}
		itemRegex, err := regexp.Compile(item)
		if err == nil {
			if itemRegex.MatchString(action) {
				return true
			}
		}
	}
	return false
}

func createActionHandlers(route Route, actions []string) []*actionHandler {
	// before we create the endpoints, lets go further and then filter by aliases.
	// There are two scenarios:
	// Scenario 1: A user would want all their actions to be endpoints. MappingPolicy -> all
	// Scenario 2: A would want to specify the actions that needs endpoint. MappingPolicy -> restrict

	settingsRoutePath := route.Path
	routePath := "/"
	if settingsRoutePath != "" {
		routePath = settingsRoutePath
	}

	settingsMappingPolicy := route.MappingPolicy
	mappingPolicy := MappingPolicyAll

	if settingsMappingPolicy == MappingPolicyAll || settingsMappingPolicy == MappingPolicyRestrict {
		mappingPolicy = settingsMappingPolicy
	}

	aliases := route.Aliases

	actionToAlias := invertStringMap(aliases)

	handlers := []*actionHandler{}
	for _, action := range actions {
		actionAlias, exists := actionToAlias[action]

		// if policy is restrict and that action is not in the list of aliases, then we skip it.
		if !exists && mappingPolicy == MappingPolicyRestrict {
			continue
		}

		handlers = append(handlers, &actionHandler{alias: actionAlias, routePath: routePath, action: action})
	}

	return handlers

}

func invertStringMap(in map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range in {
		out[value] = key
	}
	return out
}
