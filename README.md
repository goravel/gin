# Gin

Gin http driver for Goravel.

## Version

| goravel/gin | goravel/framework |
|-------------|-------------------|
| v1.0.0      | v1.13.0           |

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
