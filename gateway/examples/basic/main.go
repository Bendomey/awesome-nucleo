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
	},
}

var GatewayMixin = gateway.NewGatewayMixin(gateway.GatewayMixin{})

var GatewayApi = nucleo.ServiceSchema{
	Name: "gateway",
	Mixins: []nucleo.Mixin{
		GatewayMixin,
	},
	Settings: map[string]interface{}{
		"port": 5001,
		"routes": []gateway.Route{
			{
				Name:          "basic",
				Path:          "/api",
				MappingPolicy: gateway.MappingPolicyRestrict,
				Aliases: map[string]string{
					"POST /calculators/get": "calculator.add",
				},
				Authorization:  false,
				Authentication: false,
			},
			{
				Name:      "node-endpoints",
				Path:      "/nodes",
				Whitelist: []string{"$node.*"},
			},
		},
		// FIXME: parsing issue in nucleo-go. For now we can't pass data to function
		"onError": func() {
			// "onError": func(context *gin.Context, response nucleo.Payload) {
			// jsonSerializer := serializer.JSONSerializer{}
			// responsePayload := payload.New(map[string]interface{}{
			// 	"error": response.Error().Error(),
			// 	"type":  "NotFound",
			// 	"code":  400,
			// })
			// json := jsonSerializer.PayloadToBytes(responsePayload)

			// context.Writer.Write(json)
			fmt.Print("error occured")
		},
	},
}

func main() {
	bkr := broker.New(&nucleo.Config{LogLevel: nucleo.LogLevelInfo})

	bkr.PublishServices(GatewayApi, Calculator)

	bkr.Start()

	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, os.Interrupt, syscall.SIGTERM)

	<-signalC

	bkr.Stop()
}
