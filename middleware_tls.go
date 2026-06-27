package gin

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/unrolled/secure"
)

type TlsMiddleware struct{}

func (t *TlsMiddleware) Signature() string {
	return "goravel:tls"
}

func (t *TlsMiddleware) Handle(ctx contractshttp.Context) {
	host := ConfigFacade.GetString("http.tls.host")
	port := ConfigFacade.GetString("http.tls.port")
	cert := ConfigFacade.GetString("http.tls.ssl.cert")
	key := ConfigFacade.GetString("http.tls.ssl.key")

	if host == "" || cert == "" || key == "" || ctx.Request().Origin().TLS == nil {
		ctx.Request().Next()

		return
	}

	completeHost := host
	if port != "" {
		completeHost = host + ":" + port
	}

	secureMiddleware := secure.New(secure.Options{
		SSLRedirect: true,
		SSLHost:     completeHost,
	})

	if err := secureMiddleware.Process(ctx.Response().Writer(), ctx.Request().Origin()); err != nil {
		ctx.Request().Abort(http.StatusForbidden)
	}

	ctx.Request().Next()
}

func Tls() contractshttp.Middleware {
	return &TlsMiddleware{}
}
