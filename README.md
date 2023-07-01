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
| v1.0.2      | v1.13.0           |

## Install

1. Add package

```
go get -u github.com/goravel/gin
```

2. Register service provider, make sure it is registered first.

```
// config/app.go
import "github.com/goravel/gin"

"providers": []foundation.ServiceProvider{
    &gin.ServiceProvider{},
    ...
}
```

3. Add gin config to `config/http.go` file

```
// config/http.go
import (
    ginfacades "github.com/goravel/gin/facades"
)

"driver": "gin",

"drivers": map[string]any{
    ...
    "gin": map[string]any{
        "http": func() (http.Context, error) {
            return ginfacades.Http(), nil
        },
        "route": func() (route.Engine, error) {
            return ginfacades.Route(), nil
        },
    },
}
```

## Testing

Run command below to run test:

```
go test ./...
```
