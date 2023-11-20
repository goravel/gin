package gin

import (
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin/render"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
	configmocks "github.com/goravel/framework/mocks/config"
	"github.com/stretchr/testify/assert"
)

func TestFallback(t *testing.T) {
	mockConfig := &configmocks.Config{}
	mockConfig.On("GetBool", "app.debug").Return(true).Once()
	mockConfig.On("GetInt", "http.drivers.gin.body_limit", 4).Return(4).Once()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/fallback", nil)

	gin, err := NewRoute(mockConfig, nil)
	assert.Nil(t, err)

	gin.Fallback(func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().String(404, "not found")
	})

	gin.ServeHTTP(w, req)

	assert.Equal(t, "not found", w.Body.String())
	assert.Equal(t, http.StatusNotFound, w.Code)

	mockConfig.AssertExpectations(t)
}

func TestRun(t *testing.T) {
	var (
		err        error
		mockConfig *configmocks.Config
		route      *Route
	)

	tests := []struct {
		name        string
		setup       func(host string, port string) error
		host        string
		port        string
		expectError error
	}{
		{
			name: "error when default host is empty",
			setup: func(host string, port string) error {
				mockConfig.On("GetString", "http.host").Return(host).Once()

				go func() {
					assert.EqualError(t, route.Run(), "host can't be empty")
				}()
				time.Sleep(1 * time.Second)

				return errors.New("error")
			},
		},
		{
			name: "error when default port is empty",
			setup: func(host string, port string) error {
				mockConfig.On("GetString", "http.host").Return(host).Once()
				mockConfig.On("GetString", "http.port").Return(port).Once()

				go func() {
					assert.EqualError(t, route.Run(), "port can't be empty")
				}()
				time.Sleep(1 * time.Second)

				return errors.New("error")
			},
			host: "127.0.0.1",
		},
		{
			name: "use default host",
			setup: func(host string, port string) error {
				mockConfig.On("GetBool", "app.debug").Return(true).Once()
				mockConfig.On("GetString", "http.host").Return(host).Once()
				mockConfig.On("GetString", "http.port").Return(port).Once()

				go func() {
					assert.Nil(t, route.Run())
				}()

				time.Sleep(1 * time.Second)

				return nil
			},
			host: "127.0.0.1",
			port: "3031",
		},
		{
			name: "use custom host",
			setup: func(host string, port string) error {
				mockConfig.On("GetBool", "app.debug").Return(true).Once()

				go func() {
					assert.Nil(t, route.Run(host))
				}()

				return nil
			},
			host: "127.0.0.1:3032",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockConfig = &configmocks.Config{}
			mockConfig.On("GetBool", "app.debug").Return(true).Once()
			mockConfig.On("GetInt", "http.drivers.gin.body_limit", 4).Return(4).Once()

			route, err = NewRoute(mockConfig, nil)
			assert.Nil(t, err)
			route.Get("/", func(ctx contractshttp.Context) contractshttp.Response {
				return ctx.Response().Json(200, contractshttp.Json{
					"Hello": "Goravel",
				})
			})
			if err := test.setup(test.host, test.port); err == nil {
				time.Sleep(1 * time.Second)
				hostUrl := "http://" + test.host
				if test.port != "" {
					hostUrl = hostUrl + ":" + test.port
				}
				resp, err := http.Get(hostUrl)
				assert.Nil(t, err)
				defer resp.Body.Close()

				body, err := io.ReadAll(resp.Body)
				assert.Nil(t, err)
				assert.Equal(t, "{\"Hello\":\"Goravel\"}", string(body))
			}
			mockConfig.AssertExpectations(t)
		})
	}
}

func TestRunTLS(t *testing.T) {
	var (
		err        error
		mockConfig *configmocks.Config
		route      *Route
	)

	tests := []struct {
		name        string
		setup       func(host string, port string) error
		host        string
		port        string
		expectError error
	}{
		{
			name: "error when default host is empty",
			setup: func(host string, port string) error {
				mockConfig.On("GetString", "http.tls.host").Return(host).Once()

				go func() {
					assert.EqualError(t, route.RunTLS(), "host can't be empty")
				}()
				time.Sleep(1 * time.Second)

				return errors.New("error")
			},
		},
		{
			name: "error when default port is empty",
			setup: func(host string, port string) error {
				mockConfig.On("GetString", "http.tls.host").Return(host).Once()
				mockConfig.On("GetString", "http.tls.port").Return(port).Once()

				go func() {
					assert.EqualError(t, route.RunTLS(), "port can't be empty")
				}()
				time.Sleep(1 * time.Second)

				return errors.New("error")
			},
			host: "127.0.0.1",
		},
		{
			name: "use default host",
			setup: func(host string, port string) error {
				mockConfig.On("GetBool", "app.debug").Return(true).Once()
				mockConfig.On("GetString", "http.tls.host").Return(host).Once()
				mockConfig.On("GetString", "http.tls.port").Return(port).Once()
				mockConfig.On("GetString", "http.tls.ssl.cert").Return("test_ca.crt").Once()
				mockConfig.On("GetString", "http.tls.ssl.key").Return("test_ca.key").Once()

				go func() {
					assert.Nil(t, route.RunTLS())
				}()

				return nil
			},
			host: "127.0.0.1",
			port: "3003",
		},
		{
			name: "use custom host",
			setup: func(host string, port string) error {
				mockConfig.On("GetBool", "app.debug").Return(true).Once()
				mockConfig.On("GetString", "http.tls.ssl.cert").Return("test_ca.crt").Once()
				mockConfig.On("GetString", "http.tls.ssl.key").Return("test_ca.key").Once()

				go func() {
					assert.Nil(t, route.RunTLS(host))
				}()

				return nil
			},
			host: "127.0.0.1:3004",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockConfig = &configmocks.Config{}
			mockConfig.On("GetBool", "app.debug").Return(true).Once()
			mockConfig.On("GetInt", "http.drivers.gin.body_limit", 4).Return(4).Once()

			route, err = NewRoute(mockConfig, nil)
			assert.Nil(t, err)
			route.Get("/", func(ctx contractshttp.Context) contractshttp.Response {
				return ctx.Response().Json(200, contractshttp.Json{
					"Hello": "Goravel",
				})
			})
			if err := test.setup(test.host, test.port); err == nil {
				time.Sleep(1 * time.Second)
				tr := &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
				client := &http.Client{Transport: tr}
				hostUrl := "https://" + test.host
				if test.port != "" {
					hostUrl = hostUrl + ":" + test.port
				}
				resp, err := client.Get(hostUrl)
				assert.Nil(t, err)
				defer resp.Body.Close()

				body, err := io.ReadAll(resp.Body)
				assert.Nil(t, err)
				assert.Equal(t, "{\"Hello\":\"Goravel\"}", string(body))
			}
			mockConfig.AssertExpectations(t)
		})
	}
}

func TestRunTLSWithCert(t *testing.T) {
	var (
		err        error
		mockConfig *configmocks.Config
		route      *Route
	)

	tests := []struct {
		name        string
		setup       func(host string) error
		host        string
		expectError error
	}{
		{
			name: "error when default host is empty",
			setup: func(host string) error {
				go func() {
					assert.EqualError(t, route.RunTLSWithCert(host, "test_ca.crt", "test_ca.key"), "host can't be empty")
				}()
				time.Sleep(1 * time.Second)

				return errors.New("error")
			},
		},
		{
			name: "use default host",
			setup: func(host string) error {
				mockConfig.On("GetBool", "app.debug").Return(true).Once()

				go func() {
					assert.Nil(t, route.RunTLSWithCert(host, "test_ca.crt", "test_ca.key"))
				}()

				return nil
			},
			host: "127.0.0.1:3005",
		},
		{
			name: "use custom host",
			setup: func(host string) error {
				mockConfig.On("GetBool", "app.debug").Return(true).Once()

				go func() {
					assert.Nil(t, route.RunTLSWithCert(host, "test_ca.crt", "test_ca.key"))
				}()

				return nil
			},
			host: "127.0.0.1:3006",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockConfig = &configmocks.Config{}
			mockConfig.On("GetBool", "app.debug").Return(true).Once()
			mockConfig.On("GetInt", "http.drivers.gin.body_limit", 4).Return(4).Once()

			route, err = NewRoute(mockConfig, nil)
			assert.Nil(t, err)
			route.Get("/", func(ctx contractshttp.Context) contractshttp.Response {
				return ctx.Response().Json(200, contractshttp.Json{
					"Hello": "Goravel",
				})
			})
			if err := test.setup(test.host); err == nil {
				time.Sleep(1 * time.Second)
				tr := &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
				client := &http.Client{Transport: tr}
				resp, err := client.Get("https://" + test.host)
				assert.Nil(t, err)
				defer resp.Body.Close()

				body, err := io.ReadAll(resp.Body)
				assert.Nil(t, err)
				assert.Equal(t, "{\"Hello\":\"Goravel\"}", string(body))
			}
			mockConfig.AssertExpectations(t)
		})
	}
}

func TestNewRoute(t *testing.T) {
	var mockConfig *configmocks.Config
	defaultTemplate, err := DefaultTemplate()
	assert.Error(t, err)

	tests := []struct {
		name             string
		parameters       map[string]any
		setup            func()
		expectHTMLRender render.HTMLRender
		expectError      error
	}{
		{
			name:  "parameters is nil",
			setup: func() {},
		},
		{
			name:       "template is instance",
			parameters: map[string]any{"driver": "gin"},
			setup: func() {
				mockConfig.On("Get", "http.drivers.gin.template").Return(defaultTemplate).Once()
			},
			expectHTMLRender: defaultTemplate,
		},
		{
			name:       "template is callback and returns success",
			parameters: map[string]any{"driver": "gin"},
			setup: func() {
				mockConfig.On("Get", "http.drivers.gin.template").Return(func() (render.HTMLRender, error) {
					return defaultTemplate, nil
				}).Twice()
			},
			expectHTMLRender: defaultTemplate,
		},
		{
			name:       "template is callback and returns error",
			parameters: map[string]any{"driver": "gin"},
			setup: func() {
				mockConfig.On("Get", "http.drivers.gin.template").Return(func() (render.HTMLRender, error) {
					return nil, errors.New("error")
				}).Twice()
			},
			expectError: errors.New("error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockConfig = configmocks.NewConfig(t)
			mockConfig.On("GetBool", "app.debug").Return(true).Once()
			mockConfig.On("GetInt", "http.drivers.gin.body_limit", 4).Return(4).Once()
			test.setup()
			route, err := NewRoute(mockConfig, test.parameters)
			assert.Equal(t, test.expectError, err)
			if route != nil {
				assert.Equal(t, test.expectHTMLRender, route.instance.HTMLRender)
			}

			mockConfig.AssertExpectations(t)
		})
	}
}

type CreateUser struct {
	Name string `form:"name" json:"name" filter:"trim"`
}

func (r *CreateUser) Authorize(ctx contractshttp.Context) error {
	return nil
}

func (r *CreateUser) Rules(ctx contractshttp.Context) map[string]string {
	return map[string]string{
		"name": "required",
	}
}

func (r *CreateUser) Messages(ctx contractshttp.Context) map[string]string {
	return map[string]string{}
}

func (r *CreateUser) Attributes(ctx contractshttp.Context) map[string]string {
	return map[string]string{}
}

func (r *CreateUser) PrepareForValidation(ctx contractshttp.Context, data validation.Data) error {
	if name, exist := data.Get("name"); exist {
		return data.Set("name", name.(string)+"1")
	}

	return nil
}

type Unauthorize struct {
	Name string `form:"name" json:"name"`
}

func (r *Unauthorize) Authorize(ctx contractshttp.Context) error {
	return errors.New("error")
}

func (r *Unauthorize) Rules(ctx contractshttp.Context) map[string]string {
	return map[string]string{
		"name": "required",
	}
}

func (r *Unauthorize) Messages(ctx contractshttp.Context) map[string]string {
	return map[string]string{}
}

func (r *Unauthorize) Attributes(ctx contractshttp.Context) map[string]string {
	return map[string]string{}
}

func (r *Unauthorize) PrepareForValidation(ctx contractshttp.Context, data validation.Data) error {
	return nil
}
