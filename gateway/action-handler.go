package gateway

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Bendomey/nucleo-go"
	"github.com/Bendomey/nucleo-go/payload"
	"github.com/Bendomey/nucleo-go/serializer"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var jsonSerializer = serializer.CreateJSONSerializer(log.WithFields(log.Fields{
	"gateway": "json-serializer",
}))

type actionHandler struct {
	alias                string
	action               string
	context              nucleo.Context
	settings             map[string]interface{}
	acceptedMethodsCache map[string]bool
	route                Route
	router               *gin.RouterGroup
	authenticate         *AuthenticateMethodsFunc
	authorize            *AuthorizeMethodFunc
}

// aliasPath return the alias path(endpoint), if one exists for the action.
func (handler *actionHandler) aliasPath() string {
	if handler.alias != "" {
		parts := strings.Split(strings.TrimSpace(handler.alias), " ")
		alias := ""
		if len(parts) == 1 {
			alias = parts[0]
		} else if len(parts) == 2 {
			alias = parts[1]
		} else {
			panic(fmt.Sprint("Invalid alias format! -> ", handler.alias))
		}
		return alias
	}
	return ""
}

// pattern return the path pattern used to map URL in the http.ServeMux
func (handler *actionHandler) getFullPath() string {
	actionPath := strings.Replace(handler.action, ".", "/", -1)
	fullPath := ""
	aliasPath := handler.aliasPath()
	if aliasPath != "" {
		fullPath = fmt.Sprint("/", aliasPath)
	} else {
		fullPath = fmt.Sprint("/", actionPath)
	}
	return strings.Replace(fullPath, "//", "/", -1)
}

func (handler *actionHandler) Handler() gin.HandlerFunc {
	logger := handler.context.Logger()

	return func(ctx *gin.Context) {

		if handler.route.OnBeforeCall != nil {
			(*handler.route.OnBeforeCall)(handler.context, ctx, handler.route, handler.alias)
		}

		// Authentication call
		if handler.route.Authentication && handler.authenticate != nil {
			user := (*handler.authenticate)(handler.context, ctx, handler.alias)
			if user != nil {
				handler.context.Logger().Debug("Authenticated user", user)
				handler.context.Meta().AddMany(map[string]interface{}{
					"user": user,
				})
			} else {
				// Anonymous user
				handler.context.Logger().Debug("Anonymous user")
				handler.context.Meta().AddMany(map[string]interface{}{
					"user": nil,
				})
			}
		}

		// Authorization call
		if handler.route.Authorization && handler.authorize != nil {
			(*handler.authorize)(handler.context, ctx, handler.alias)
		}

		logRequestFormatType, logRequestFormatTypeExists := handler.settings["logRequest"].(nucleo.LogLevelType)
		if logRequestFormatTypeExists {
			logRequestLogger := getLogger(logRequestFormatType, logger)
			logRequestLogger("Call '", handler.action, "' action")
		}

		params := paramsFromRequest(ctx.Request, logger)

		logRequestParamsFormatType, logRequestParamsFormatTypeExists := handler.settings["logRequestParams"].(nucleo.LogLevelType)
		if logRequestParamsFormatTypeExists {
			logRequestParamsLogger := getLogger(logRequestParamsFormatType, logger)
			logRequestParamsLogger("Params: ", params)
		}

		callActionResponse := <-handler.context.Call(handler.action, params)

		logResponseDataFormatType, logResponseDataFormatTypeExists := handler.settings["logResponseData"].(nucleo.LogLevelType)
		if logResponseDataFormatTypeExists {
			logResponseDataLogger := getLogger(logResponseDataFormatType, logger)
			logResponseDataLogger("Data: ", callActionResponse)
		}

		if handler.route.OnAfterCall != nil {
			(*handler.route.OnAfterCall)(handler.context, ctx, handler.route, callActionResponse)
		}

		handler.sendReponse(logger, callActionResponse, ctx)

	}
}

var succesStatusCode = 200
var errorStatusCode = 500

func (handler *actionHandler) responesErrorHandler(ginContext *gin.Context, result nucleo.Payload) {
	logger := handler.context.Logger()

	ginContext.Writer.WriteHeader(errorStatusCode)

	log4XXResponses, log4XXResponsesExists := handler.settings["log4XXResponses"].(nucleo.LogLevelType)
	if log4XXResponsesExists {
		log4XXResponsesLogger := getLogger(log4XXResponses, logger)
		log4XXResponsesLogger("Gateway  Request error! - action: ", handler.action, " error ", result.Error())
	}

	onError, onErrorExists := handler.settings["onError"].(func(context *gin.Context, response nucleo.Payload))

	// if user has onError middleware configured, they will be ablle to override it.
	if onErrorExists {
		// FIXME: parsing issue in nucleo-go
		onError(ginContext, result)
	} else {
		// return response.
		json := jsonSerializer.PayloadToBytes(payload.Empty().Add("error", result.Error().Error()))
		ginContext.Writer.Write(json)
	}
}

// sendReponse send the result payload  back using the ResponseWriter
func (handler *actionHandler) sendReponse(logger *log.Entry, result nucleo.Payload, ginContext *gin.Context) {
	// Return with a JSON object
	ginContext.Writer.Header().Add("Content-Type", "application/json; charset=utf-8")

	if result.IsError() {
		handler.responesErrorHandler(ginContext, result)
		return
	}

	var json []byte
	ginContext.Writer.WriteHeader(succesStatusCode)
	json = jsonSerializer.PayloadToBytes(result)

	logger.Debug("Gateway SendReponse() - action: ", handler.action, " json: ", string(json))
	ginContext.Writer.Write(json)
}

// acceptedMethods return a map of accepted methods for this handler.
func (handler *actionHandler) AcceptedMethods() map[string]bool {
	if handler.acceptedMethodsCache != nil {
		return handler.acceptedMethodsCache
	}
	if handler.alias != "" {
		parts := strings.Split(strings.TrimSpace(handler.alias), " ")
		if len(parts) == 2 {
			method := strings.ToUpper(parts[0])
			if validMethod(method) {
				handler.acceptedMethodsCache = map[string]bool{
					method: true,
				}
				return handler.acceptedMethodsCache
			}
		}
	}
	handler.acceptedMethodsCache = map[string]bool{
		"GET":    true,
		"POST":   true,
		"PUT":    true,
		"PATCH":  true,
		"DELETE": true,
	}
	return handler.acceptedMethodsCache
}

var validMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

func validMethod(method string) bool {
	for _, item := range validMethods {
		if item == method {
			return true
		}
	}
	return false
}

func paramsFromRequestForm(request *http.Request, logger *log.Entry) (map[string]interface{}, error) {
	params := map[string]interface{}{}
	err := request.ParseForm()
	if err != nil {
		logger.Errorln("Error calling request.ParseForm() -> ", err)
		return nil, err
	}
	for name, value := range request.Form {
		if len(value) == 1 {
			params[name] = value[0]
		} else {
			params[name] = value
		}
	}
	return params, nil
}

// paramsFromRequest extract params from body and URL into a payload.
func paramsFromRequest(request *http.Request, logger *log.Entry) nucleo.Payload {
	mvalues, err := paramsFromRequestForm(request, logger)
	if len(mvalues) > 0 {
		return payload.New(mvalues)
	}
	if err != nil {
		return payload.Error("Error trying to parse request form values. Error: ", err.Error())
	}

	bts, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return payload.Error("Error trying to parse request body. Error: ", err.Error())
	}
	return jsonSerializer.BytesToPayload(&bts)
}

func getLogger(logType nucleo.LogLevelType, existingLogger *log.Entry) func(args ...interface{}) {

	if logType == nucleo.LogLevelWarn {
		return existingLogger.Warnln
	} else if logType == nucleo.LogLevelDebug {
		return existingLogger.Debugln
	} else if logType == nucleo.LogLevelTrace {
		return existingLogger.Traceln
	} else if logType == nucleo.LogLevelError {
		return existingLogger.Errorln
	} else if logType == nucleo.LogLevelFatal {
		return existingLogger.Fatalln
	} else if logType == nucleo.LogLevelInfo {
		return existingLogger.Infoln
	}

	// don't log when it's not set
	return func(args ...interface{}) {}
}
