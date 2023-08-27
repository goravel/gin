# Gin

[![Doc](https://pkg.go.dev/badge/github.com/goravel/gin)](https://pkg.go.dev/github.com/goravel/gin)
[![Go](https://img.shields.io/github/go-mod/go-version/goravel/gin)](https://go.dev/)
[![Release](https://img.shields.io/github/release/goravel/gin.svg)](https://github.com/goravel/gin/releases)
[![Test](https://github.com/goravel/gin/actions/workflows/test.yml/badge.svg)](https://github.com/goravel/gin/actions)
[![Report Card](https://goreportcard.com/badge/github.com/goravel/gin)](https://goreportcard.com/report/github.com/goravel/gin)
[![Codecov](https://codecov.io/gh/goravel/gin/branch/master/graph/badge.svg)](https://codecov.io/gh/goravel/gin)
![License](https://img.shields.io/github/license/goravel/gin)

Gin http driver for Goravel.

## Version

| goravel/gin | goravel/framework |
|-------------|-------------------|
| v1.1.x      | v1.13.x           |

## Install

1. Add package

```
go get -u github.com/goravel/gin
```

2. Register service provider

```
// config/app.go
import "github.com/goravel/gin"

"providers": []foundation.ServiceProvider{
    ...
    &gin.ServiceProvider{},
}
```

3. Add gin config to `config/http.go` file

```
// config/http.go
import (
    ginfacades "github.com/goravel/gin/facades"
    "github.com/gin-gonic/gin/render"
    "github.com/goravel/gin"
)

"default": "gin",

"drivers": map[string]any{
    "gin": map[string]any{
        "route": func() (route.Engine, error) {
            return ginfacades.Route(), nil
        },
        // Optional, default is http/template
        "template": func() (render.HTMLRender, error) {
            return gin.DefaultTemplate(), nil
        },
    },
},
```

## Testing

Run command below to run test:

```
go test ./...
```
