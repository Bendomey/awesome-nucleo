# Gateway
The nucleo-gateway is the official API gateway service for Nucleo framework. Use it to publish your services

## Features
- [ ] support HTTP & HTTPS
- [ ] serve static files
- [x] multiple routes
- [x] support Connect-like middlewares in global-level, route-level and alias-level.
- [x] alias names (with named parameters & REST routes)
- [x] whitelist
- [x] multiple body parsers (json, urlencoded). Use our middlewares to set this up
- [x] [CORS headers](https://github.com/gin-contrib/cors). Use with our middlewares.
- [x] Rate limiter. Use custom middleware for this.
- [x] before & after call hooks
- [ ] Buffer & Stream handling
- [x] support authentication and authorization

## Installation
```bash
go get github.com/Bendomey/awesome-nucleo/gateway
```

## Usage
```go
import (
    "github.com/Bendomey/nucleo-go"
    "github.com/Bendomey/awesome-nucleo/gateway"
)

var GatewayMixin = gateway.NewGatewayMixin(gateway.GatewayMixin{})

var ApiService = nucleo.ServiceSchema{
    Name: "api",
    Mixins:  []nucleo.Mixin{GatewayMixin},
    Settings: map[string]interface{}{
        "port": 5001,
    }
}
```


## License
awesome-nucleo is available under the [Apache License](https://www.tldrlegal.com/license/apache-license-2-0-apache-2-0)

## Contact
Copyright (c) 2023 Awesome Nucleo

[![@awesome-nucleo](https://img.shields.io/badge/github-nucleo-green.svg)](https://github.com/Bendomey/awesome-nucleo)