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
	routePath            string
	action               string
	context              nucleo.Context
	acceptedMethodsCache map[string]bool
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
		fullPath = fmt.Sprint(handler.routePath, "/", aliasPath)
	} else {
		fullPath = fmt.Sprint(handler.routePath, "/", actionPath)
	}
	return strings.Replace(fullPath, "//", "/", -1)
}

func (handler *actionHandler) Handler() gin.HandlerFunc {
	logger := handler.context.Logger()

	return func(ctx *gin.Context) {
		handler.sendReponse(logger, <-handler.context.Call(handler.action, paramsFromRequest(ctx.Request, logger)), ctx.Writer)
	}
}

var succesStatusCode = 200
var errorStatusCode = 500

// sendReponse send the result payload  back using the ResponseWriter
func (handler *actionHandler) sendReponse(logger *log.Entry, result nucleo.Payload, response gin.ResponseWriter) {
	var json []byte
	response.Header().Add("Content-Type", "application/json")
	if result.IsError() {
		response.WriteHeader(errorStatusCode)
		json = jsonSerializer.PayloadToBytes(payload.Empty().Add("error", result.Error().Error()))
	} else {
		response.WriteHeader(succesStatusCode)
		json = jsonSerializer.PayloadToBytes(result)
	}
	logger.Debug("Gateway SendReponse() - action: ", handler.action, " json: ", string(json), " result.IsError(): ", result.IsError())
	response.Write(json)
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
		logger.Error("Error calling request.ParseForm() -> ", err)
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
