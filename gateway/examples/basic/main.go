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
	print("authenticate called")

	return map[string]interface{}{
		"name":  "Benjamin",
		"token": "ds.kljvbajh akjsbfkhas fjnas",
	}
}

var authorizeHandler = func(context nucleo.Context, ginContext *gin.Context, alias string) {
	print("authorize called")
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
					"POST /calculators/get":  "calculator.add",
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
			responsePayload := payload.New(map[string]interface{}{
				"error": response.Error().Error(),
				"type":  "NotFound",
				"code":  400,
			})
			json := jsonSerializer.PayloadToBytes(responsePayload)

			context.Writer.Write(json)
			fmt.Print("error occured")
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
