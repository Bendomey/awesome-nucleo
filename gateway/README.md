# Gateway
The nucleo-gateway is the official API gateway service for Nucleo framework. Use it to publish your services

## Features
- [ ] support HTTP & HTTPS
- [ ] serve static files
- [ ] multiple routes
- [ ] support Connect-like middlewares in global-level, route-level and alias-level.
- [ ] alias names (with named parameters & REST routes)
- [ ] whitelist
- [ ] multiple body parsers (json, urlencoded)
- [ ] CORS headers
- [ ] Rate limiter
- [ ] before & after call hooks
- [ ] Buffer & Stream handling
- [ ] support authorization

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

var ApiService = nucleo.ServiceSchema{
    Name: "api",
    Mixins:  []nucleo.Mixin{gateway.GatewayService}
    Settings: map[string]interface{}{}
}

```


## License
awesome-nucleo is available under the [Apache License](https://www.tldrlegal.com/license/apache-license-2-0-apache-2-0)

## Contact
Copyright (c) 2023 Awesome Nucleo

[![@awesome-nucleo](https://img.shields.io/badge/github-nucleo-green.svg)](https://github.com/Bendomey/awesome-nucleo)