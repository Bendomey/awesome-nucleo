package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Bendomey/awesome-nucleo/gateway"
	"github.com/Bendomey/nucleo-go"
	"github.com/Bendomey/nucleo-go/broker"
	nucleoError "github.com/Bendomey/nucleo-go/errors"
	"github.com/Bendomey/nucleo-go/payload"
	"github.com/Bendomey/nucleo-go/serializer"
	"github.com/gin-gonic/gin"
)

var Calculator = nucleo.ServiceSchema{
	Name:     "calculator",
	Settings: map[string]interface{}{},
	Actions: []nucleo.Action{
		{
			Name:        "add",
			Description: "add two numbers",
			Params: map[string]interface{}{
				"a": "number",
				"b": "number",
			},
			Handler: func(ctx nucleo.Context, params nucleo.Payload) interface{} {
				return errors.New("hello world")
				// return params.Get("a").Int() + params.Get("b").Int()
			},
		},
		{
			Name:        "hello",
			Description: "print hello wotld",
			Handler: func(ctx nucleo.Context, params nucleo.Payload) interface{} {
				return "hello world"
			},
		},
	},
}
var authenticateHandler = func(context nucleo.Context, ginContext *gin.Context, alias string) interface{} {
	fmt.Println("authenticate called")

	return map[string]interface{}{
		"name":  "Benjamin",
		"token": "ds.kljvbajh akjsbfkhas fjnas",
	}
}

var authorizeHandler = func(context nucleo.Context, ginContext *gin.Context, alias string) {
	fmt.Println("authorize called")
}

var GatewayMixin = gateway.NewGatewayMixin(gateway.GatewayMixin{
	Authenticate: &authenticateHandler,
	Authorize:    &authorizeHandler,
})

var OnBeforeCallHandler = func(context nucleo.Context, ginContext *gin.Context, route gateway.Route, alias string) {
	fmt.Println("on before call handler")
}

var OnAfterCallHandler = func(context nucleo.Context, ginContext *gin.Context, route gateway.Route, response nucleo.Payload) {
	fmt.Println("on after call handler")
}

var GatewayApi = nucleo.ServiceSchema{
	Name: "gateway",
	Mixins: []nucleo.Mixin{
		GatewayMixin,
	},
	Settings: map[string]interface{}{
		"port": 5001,
		"use":  []gin.HandlerFunc{},
		"routes": []gateway.Route{
			{
				Name:      "node-endpoints",
				Path:      "/nodes",
				Whitelist: []string{"$node.*"},
				Use:       []gin.HandlerFunc{},
			},
			{
				Name:          "basic",
				Path:          "/api",
				Use:           []gin.HandlerFunc{},
				MappingPolicy: gateway.MappingPolicyRestrict,
				Aliases: map[string]string{
					"POST /calculators":      "calculator.add",
					"GET /calculators/hello": "calculator.hello",
				},
				OnBeforeCall:   &OnBeforeCallHandler,
				OnAfterCall:    &OnAfterCallHandler,
				Authorization:  true,
				Authentication: true,
			},
		},
		"onError": func(context *gin.Context, response nucleo.Payload) {
			jsonSerializer := serializer.JSONSerializer{}

			nucleoError := response.Value().(nucleoError.NucleoError)

			responsePayload := payload.New(map[string]interface{}{
				"message": nucleoError.Message,
				"type":    nucleoError.Type,
				"code":    nucleoError.Code,
				"data":    nucleoError.Data,
			})
			json := jsonSerializer.PayloadToBytes(responsePayload)

			context.Writer.WriteHeader(nucleoError.Code)
			context.Writer.Write(json)

		},
	},
}

func main() {
	bkr := broker.New(&nucleo.Config{LogLevel: nucleo.LogLevelDebug})

	bkr.PublishServices(GatewayApi, Calculator)

	bkr.Start()

	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, os.Interrupt, syscall.SIGTERM)

	<-signalC

	bkr.Stop()
}
