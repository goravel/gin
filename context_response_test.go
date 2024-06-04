package gin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	contractshttp "github.com/goravel/framework/contracts/http"
	configmocks "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/support/json"
	"github.com/stretchr/testify/assert"
)

func TestResponse(t *testing.T) {
	var (
		err        error
		gin        *Route
		req        *http.Request
		mockConfig *configmocks.Config
	)
	beforeEach := func() {
		mockConfig = &configmocks.Config{}
		mockConfig.On("GetBool", "app.debug").Return(true).Once()
		mockConfig.On("GetInt", "http.drivers.gin.body_limit", 4096).Return(4096).Once()
	}
	tests := []struct {
		name              string
		method            string
		url               string
		cookieName        string
		setup             func(method, url string) error
		expectCode        int
		expectBody        string
		expectHeader      string
		expectCookieValue string
	}{
		{
			name:   "Data",
			method: "GET",
			url:    "/data",
			setup: func(method, url string) error {
				gin.Get("/data", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Data(http.StatusOK, "text/html; charset=utf-8", []byte("<b>Goravel</b>"))
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "<b>Goravel</b>",
		},
		{
			name:   "Success Data",
			method: "GET",
			url:    "/success/data",
			setup: func(method, url string) error {
				gin.Get("/success/data", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Data("text/html; charset=utf-8", []byte("<b>Goravel</b>"))
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "<b>Goravel</b>",
		},
		{
			name:   "Json",
			method: "GET",
			url:    "/json",
			setup: func(method, url string) error {
				gin.Get("/json", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Json(http.StatusOK, contractshttp.Json{
						"id": "1",
					})
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name:   "String",
			method: "GET",
			url:    "/string",
			setup: func(method, url string) error {
				gin.Get("/string", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().String(http.StatusCreated, "Goravel")
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusCreated,
			expectBody: "Goravel",
		},
		{
			name:   "Success Json",
			method: "GET",
			url:    "/success/json",
			setup: func(method, url string) error {
				gin.Get("/success/json", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"id": "1",
					})
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name:   "Success String",
			method: "GET",
			url:    "/success/string",
			setup: func(method, url string) error {
				gin.Get("/success/string", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().String("Goravel")
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "Goravel",
		},
		{
			name:   "File",
			method: "GET",
			url:    "/file",
			setup: func(method, url string) error {
				gin.Get("/file", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().File("./README.md")
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
		},
		{
			name:   "Download",
			method: "GET",
			url:    "/download",
			setup: func(method, url string) error {
				gin.Get("/download", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Download("./README.md", "README.md")
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
		},
		{
			name:   "Header",
			method: "GET",
			url:    "/header",
			setup: func(method, url string) error {
				gin.Get("/header", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Header("Hello", "goravel").String(http.StatusOK, "Goravel")
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode:   http.StatusOK,
			expectBody:   "Goravel",
			expectHeader: "goravel",
		},
		{
			name:   "NoContent",
			method: "GET",
			url:    "/no/content",
			setup: func(method, url string) error {
				gin.Get("/no/content", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().NoContent()
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusNoContent,
		},
		{
			name:   "NoContentWithCode",
			method: "GET",
			url:    "/no/content/with/code",
			setup: func(method, url string) error {
				gin.Get("/no/content/with/code", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().NoContent(http.StatusAccepted)
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusAccepted,
		},
		{
			name:   "Origin",
			method: "GET",
			url:    "/origin",
			setup: func(method, url string) error {
				mockConfig.On("Get", "cors.paths").Return([]string{}).Once()
				mockConfig.On("GetString", "http.tls.host").Return("").Once()
				mockConfig.On("GetString", "http.tls.port").Return("").Once()
				mockConfig.On("GetString", "http.tls.ssl.cert").Return("").Once()
				mockConfig.On("GetString", "http.tls.ssl.key").Return("").Once()
				ConfigFacade = mockConfig

				gin.GlobalMiddleware(func(ctx contractshttp.Context) {
					ctx.Response().Header("global", "goravel")
					ctx.Request().Next()

					assert.Equal(t, "Goravel", ctx.Response().Origin().Body().String())
					assert.Equal(t, "goravel", ctx.Response().Origin().Header().Get("global"))
					assert.Equal(t, 7, ctx.Response().Origin().Size())
					assert.Equal(t, 200, ctx.Response().Origin().Status())
				})
				gin.Get("/origin", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().String(http.StatusOK, "Goravel")
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "Goravel",
		},
		{
			name:   "Redirect",
			method: "GET",
			url:    "/redirect",
			setup: func(method, url string) error {
				gin.Get("/redirect", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Redirect(http.StatusMovedPermanently, "https://goravel.dev")
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusMovedPermanently,
			expectBody: "<a href=\"https://goravel.dev\">Moved Permanently</a>.\n\n",
		},
		{
			name:       "WithoutCookie",
			method:     "GET",
			url:        "/without/cookie",
			cookieName: "goravel",
			setup: func(method, url string) error {
				gin.Get("/without/cookie", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().WithoutCookie("goravel").String(http.StatusOK, "Goravel")
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}
				req.AddCookie(&http.Cookie{
					Name:  "goravel",
					Value: "goravel",
				})

				return nil
			},
			expectCode:        http.StatusOK,
			expectBody:        "Goravel",
			expectCookieValue: "",
		},
		{
			name:       "Cookie",
			method:     "GET",
			url:        "/cookie",
			cookieName: "goravel",
			setup: func(method, url string) error {
				gin.Get("/cookie", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Cookie(contractshttp.Cookie{
						Name:  "goravel",
						Value: "goravel",
					}).String(http.StatusOK, "Goravel")
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode:        http.StatusOK,
			expectBody:        "Goravel",
			expectCookieValue: "goravel",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			gin, err = NewRoute(mockConfig, nil)
			assert.Nil(t, err)

			assert.Nil(t, test.setup(test.method, test.url))

			w := httptest.NewRecorder()

			gin.ServeHTTP(w, req)

			if test.expectBody != "" {
				assert.Equal(t, test.expectBody, w.Body.String(), test.name)
			}
			if test.expectHeader != "" {
				assert.Equal(t, test.expectHeader, strings.Join(w.Header().Values("Hello"), ""), test.name)
			}
			if test.cookieName != "" {
				cookies := w.Result().Cookies()
				exist := false
				for _, cookie := range cookies {
					if cookie.Name == test.cookieName {
						exist = true
						assert.Equal(t, test.expectCookieValue, cookie.Value)
					}
				}
				assert.True(t, exist)
			}
			assert.Equal(t, test.expectCode, w.Code, test.name)

			mockConfig.AssertExpectations(t)
		})
	}
}

func TestResponse_Success(t *testing.T) {
	var (
		err        error
		route      *Route
		req        *http.Request
		mockConfig *configmocks.Config
	)
	beforeEach := func() {
		mockConfig = &configmocks.Config{}
		mockConfig.On("GetBool", "app.debug").Return(false).Once()
		mockConfig.On("GetInt", "http.drivers.gin.body_limit", 4096).Return(4096).Once()
		ConfigFacade = mockConfig
	}
	tests := []struct {
		name           string
		method         string
		url            string
		setup          func(method, url string) error
		expectCode     int
		expectBody     string
		expectBodyJson string
	}{
		{
			name:   "Data",
			method: "GET",
			url:    "/data",
			setup: func(method, url string) error {
				route.Get("/data", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Data("text/html; charset=utf-8", []byte("<b>Goravel</b>"))
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "<b>Goravel</b>",
		},
		{
			name:   "Json",
			method: "GET",
			url:    "/json",
			setup: func(method, url string) error {
				route.Get("/json", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"id": "1",
					})
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode:     http.StatusOK,
			expectBodyJson: "{\"id\":\"1\"}",
		},
		{
			name:   "String",
			method: "GET",
			url:    "/string",
			setup: func(method, url string) error {
				route.Get("/string", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().String("Goravel")
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "Goravel",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			route, err = NewRoute(mockConfig, nil)
			assert.Nil(t, err)

			err := test.setup(test.method, test.url)
			assert.Nil(t, err)

			w := httptest.NewRecorder()
			route.ServeHTTP(w, req)

			if test.expectBody != "" {
				assert.Equal(t, test.expectBody, w.Body.String())
			}
			if test.expectBodyJson != "" {
				bodyMap := make(map[string]any)
				exceptBodyMap := make(map[string]any)

				err = json.Unmarshal(w.Body.Bytes(), &bodyMap)
				assert.Nil(t, err)
				err = json.UnmarshalString(test.expectBodyJson, &exceptBodyMap)
				assert.Nil(t, err)

				assert.Equal(t, exceptBodyMap, bodyMap)
			}

			assert.Equal(t, test.expectCode, w.Code)

			mockConfig.AssertExpectations(t)
		})
	}
}

func TestResponse_Status(t *testing.T) {
	var (
		err        error
		route      *Route
		req        *http.Request
		mockConfig *configmocks.Config
	)
	beforeEach := func() {
		mockConfig = &configmocks.Config{}
		mockConfig.On("GetBool", "app.debug").Return(false).Once()
		mockConfig.On("GetInt", "http.drivers.gin.body_limit", 4096).Return(4096).Once()
		ConfigFacade = mockConfig
	}
	tests := []struct {
		name           string
		method         string
		url            string
		setup          func(method, url string) error
		expectCode     int
		expectBody     string
		expectBodyJson string
	}{
		{
			name:   "Data",
			method: "GET",
			url:    "/data",
			setup: func(method, url string) error {
				route.Get("/data", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Status(http.StatusCreated).Data("text/html; charset=utf-8", []byte("<b>Goravel</b>"))
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusCreated,
			expectBody: "<b>Goravel</b>",
		},
		{
			name:   "Json",
			method: "GET",
			url:    "/json",
			setup: func(method, url string) error {
				route.Get("/json", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Status(http.StatusCreated).Json(contractshttp.Json{
						"id": "1",
					})
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode:     http.StatusCreated,
			expectBodyJson: "{\"id\":\"1\"}",
		},
		{
			name:   "String",
			method: "GET",
			url:    "/string",
			setup: func(method, url string) error {
				route.Get("/string", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Status(http.StatusCreated).String("Goravel")
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusCreated,
			expectBody: "Goravel",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			route, err = NewRoute(mockConfig, nil)
			assert.Nil(t, err)

			err := test.setup(test.method, test.url)
			assert.Nil(t, err)

			w := httptest.NewRecorder()
			route.ServeHTTP(w, req)

			if test.expectBody != "" {
				assert.Equal(t, test.expectBody, w.Body.String())
			}
			if test.expectBodyJson != "" {
				bodyMap := make(map[string]any)
				exceptBodyMap := make(map[string]any)

				err = json.Unmarshal(w.Body.Bytes(), &bodyMap)
				assert.Nil(t, err)
				err = json.UnmarshalString(test.expectBodyJson, &exceptBodyMap)
				assert.Nil(t, err)

				assert.Equal(t, exceptBodyMap, bodyMap)
			}

			assert.Equal(t, test.expectCode, w.Code)

			mockConfig.AssertExpectations(t)
		})
	}
}
