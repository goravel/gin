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
| v1.18.x     | v1.18.x           |
| v1.17.x     | v1.17.x           |
| v1.4.x      | v1.16.x           |
| v1.3.x      | v1.15.x           |
| v1.2.x      | v1.14.x           |
| v1.1.x      | v1.13.x           |

## Install

Run the command below in your project to install the package automatically:

```
./artisan package:install github.com/goravel/gin
```

Or check [the setup file](./setup/setup.go) to install the package manually.

## Configuration

You can define the `template` configuration. If omitted, `DefaultTemplate()` is used automatically as a fallback, which loads views from `resources/views` and any registered package views.

You can provide a custom template configuration in two forms:

- **`func() (render.HTMLRender, error)`** — a callback that returns a custom HTML renderer (e.g. to configure custom delimiters or a FuncMap).
- **`render.HTMLRender`** — a pre-built renderer instance.

**Custom example:**

```go
import (
    "html/template"

    "github.com/gin-gonic/gin/render"
    "github.com/goravel/gin"
)

"template": func() (render.HTMLRender, error) {
    return gin.NewTemplate(gin.RenderOptions{
        Delims:  &gin.Delims{Left: "{[", Right: "]}"},
        FuncMap: template.FuncMap{
            "upper": strings.ToUpper,
        },
    })
},
```

## Testing

Run command below to run test:

```
go test ./...
```
