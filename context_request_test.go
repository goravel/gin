package gin

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
	frameworkfilesystem "github.com/goravel/framework/filesystem"
	foundationjson "github.com/goravel/framework/foundation/json"
	configmocks "github.com/goravel/framework/mocks/config"
	filesystemmocks "github.com/goravel/framework/mocks/filesystem"
	logmocks "github.com/goravel/framework/mocks/log"
	validationmocks "github.com/goravel/framework/mocks/validation"
	"github.com/goravel/framework/session"
	"github.com/goravel/framework/support/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRequest(t *testing.T) {
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
		name       string
		method     string
		url        string
		setup      func(method, url string) error
		expectCode int
		expectBody string
	}{
		{
			name:   "All when Get and query is empty",
			method: "GET",
			url:    "/all",
			setup: func(method, url string) error {
				gin.Get("/all", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"all": ctx.Request().All(),
					})
				})

				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"all\":{}}",
		},
		{
			name:   "All when Get and query is not empty",
			method: "GET",
			url:    "/all?a=1&a=2&b=3",
			setup: func(method, url string) error {
				gin.Get("/all", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"all": ctx.Request().All(),
					})
				})

				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"all\":{\"a\":\"1,2\",\"b\":\"3\"}}",
		},
		{
			name:   "All with form when Post",
			method: "POST",
			url:    "/all?a=1&a=2&b=3",
			setup: func(method, url string) error {
				gin.Post("/all", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"all": ctx.Request().All(),
					})
				})

				payload := &bytes.Buffer{}
				writer := multipart.NewWriter(payload)

				if err := writer.WriteField("b", "4"); err != nil {
					return err
				}
				if err := writer.WriteField("e", "e"); err != nil {
					return err
				}

				readme, err := os.Open("./README.md")
				if err != nil {
					return err
				}
				defer readme.Close()
				part1, err := writer.CreateFormFile("file", filepath.Base("./README.md"))
				if err != nil {
					return err
				}

				if _, err = io.Copy(part1, readme); err != nil {
					return err
				}

				if err := writer.Close(); err != nil {
					return err
				}

				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", writer.FormDataContentType())

				return nil
			},
			expectCode: http.StatusOK,
		},
		{
			name:   "All with empty form when Post",
			method: "POST",
			url:    "/all?a=1&a=2&b=3",
			setup: func(method, url string) error {
				gin.Post("/all", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"all": ctx.Request().All(),
					})
				})

				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "multipart/form-data;boundary=0")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"all\":{\"a\":\"1,2\",\"b\":\"3\"}}",
		},
		{
			name:   "All with json when Post",
			method: "POST",
			url:    "/all?a=1&a=2&name=3",
			setup: func(method, url string) error {
				gin.Post("/all", func(ctx contractshttp.Context) contractshttp.Response {
					all := ctx.Request().All()
					type Test struct {
						Name string
						Age  int
					}
					var test Test
					_ = ctx.Request().Bind(&test)

					return ctx.Response().Success().Json(contractshttp.Json{
						"all":  all,
						"name": test.Name,
						"age":  test.Age,
					})
				})

				payload := strings.NewReader(`{
					"Name": "goravel",
					"Age": 1
				}`)
				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"age\":1,\"all\":{\"Age\":1,\"Name\":\"goravel\",\"a\":\"1,2\",\"name\":\"3\"},\"name\":\"goravel\"}",
		},
		{
			name:   "All with error json when Post",
			method: "POST",
			url:    "/all?a=1&a=2&name=3",
			setup: func(method, url string) error {
				mockLog := &logmocks.Log{}
				LogFacade = mockLog
				mockLog.On("Error", mock.Anything).Twice()

				gin.Post("/all", func(ctx contractshttp.Context) contractshttp.Response {
					all := ctx.Request().All()
					type Test struct {
						Name string
						Age  int
					}
					var test Test
					_ = ctx.Request().Bind(&test)

					return ctx.Response().Success().Json(contractshttp.Json{
						"all":  all,
						"name": test.Name,
						"age":  test.Age,
					})
				})

				payload := strings.NewReader(`{
					"Name": "goravel",
					"Age": 1,
				}`)
				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"age\":0,\"all\":{\"a\":\"1,2\",\"name\":\"3\"},\"name\":\"\"}",
		},
		{
			name:   "All with empty json when Post",
			method: "POST",
			url:    "/all?a=1&a=2&name=3",
			setup: func(method, url string) error {
				gin.Post("/all", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"all": ctx.Request().All(),
					})
				})

				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"all\":{\"a\":\"1,2\",\"name\":\"3\"}}",
		},
		{
			name:   "All with json when Put",
			method: "PUT",
			url:    "/all?a=1&a=2&b=3",
			setup: func(method, url string) error {
				gin.Put("/all", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"all": ctx.Request().All(),
					})
				})

				payload := strings.NewReader(`{
					"b": 4,
					"e": "e"
				}`)
				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"all\":{\"a\":\"1,2\",\"b\":4,\"e\":\"e\"}}",
		},
		{
			name:   "All with json when Delete",
			method: "DELETE",
			url:    "/all?a=1&a=2&b=3",
			setup: func(method, url string) error {
				gin.Delete("/all", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"all": ctx.Request().All(),
					})
				})

				payload := strings.NewReader(`{
					"b": 4,
					"e": "e"
				}`)
				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"all\":{\"a\":\"1,2\",\"b\":4,\"e\":\"e\"}}",
		},
		{
			name:   "Methods",
			method: "GET",
			url:    "/methods/1?name=Goravel",
			setup: func(method, url string) error {
				gin.Get("/methods/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"id":       ctx.Request().Input("id"),
						"name":     ctx.Request().Query("name", "Hello"),
						"header":   ctx.Request().Header("Hello", "World"),
						"method":   ctx.Request().Method(),
						"path":     ctx.Request().Path(),
						"url":      ctx.Request().Url(),
						"full_url": ctx.Request().FullUrl(),
						"ip":       ctx.Request().Ip(),
					})
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				req.Header.Set("Hello", "goravel")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"full_url\":\"\",\"header\":\"goravel\",\"id\":\"1\",\"ip\":\"\",\"method\":\"GET\",\"name\":\"Goravel\",\"path\":\"/methods/1\",\"url\":\"\"}",
		},
		{
			name:   "Headers",
			method: "GET",
			url:    "/headers",
			setup: func(method, url string) error {
				gin.Get("/headers", func(ctx contractshttp.Context) contractshttp.Response {
					str, _ := json.Marshal(ctx.Request().Headers())
					return ctx.Response().Success().String(string(str))
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}
				req.Header.Set("Hello", "Goravel")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"Hello\":[\"Goravel\"]}",
		},
		{
			name:   "Route",
			method: "GET",
			url:    "/route/1/2/3/a",
			setup: func(method, url string) error {
				gin.Get("/route/{string}/{int}/{int64}/{string1}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"string": ctx.Request().Route("string"),
						"int":    ctx.Request().RouteInt("int"),
						"int64":  ctx.Request().RouteInt64("int64"),
						"error":  ctx.Request().RouteInt("string1"),
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
			expectBody: "{\"error\":0,\"int\":2,\"int64\":3,\"string\":\"1\"}",
		},
		{
			name:   "Input - application/json",
			method: "POST",
			url:    "/input1/1?id=2",
			setup: func(method, url string) error {
				gin.Post("/input1/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"id":     ctx.Request().Input("id"),
						"int":    ctx.Request().Input("int"),
						"map":    ctx.Request().Input("map"),
						"string": ctx.Request().Input("string"),
					})
				})

				payload := strings.NewReader(`{
					"id": "3",
					"string": ["string 1", "string 2"],
					"int": [1, 2],
					"map": {"a": "b"}
				}`)
				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"3\",\"int\":\"1,2\",\"map\":\"{\\\"a\\\":\\\"b\\\"}\",\"string\":\"string 1,string 2\"}",
		},
		{
			name:   "Input - multipart/form-data",
			method: "POST",
			url:    "/input2/1?id=2",
			setup: func(method, url string) error {
				gin.Post("/input2/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"id":     ctx.Request().Input("id"),
						"map":    ctx.Request().Input("map"),
						"string": ctx.Request().Input("string[]"),
					})
				})

				payload := &bytes.Buffer{}
				writer := multipart.NewWriter(payload)
				if err := writer.WriteField("id", "4"); err != nil {
					return err
				}
				if err := writer.WriteField("string[]", "string 1"); err != nil {
					return err
				}
				if err := writer.WriteField("string[]", "string 2"); err != nil {
					return err
				}
				if err := writer.WriteField("map", "{\"a\":\"b\"}"); err != nil {
					return err
				}
				if err := writer.Close(); err != nil {
					return err
				}

				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", writer.FormDataContentType())

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"4\",\"map\":\"{\\\"a\\\":\\\"b\\\"}\",\"string\":\"string 1,string 2\"}",
		},
		{
			name:   "Input - application/x-www-form-urlencoded",
			method: "POST",
			url:    "/input/url/1?id=2",
			setup: func(method, url string) error {
				gin.Post("/input/url/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"id":     ctx.Request().Input("id"),
						"map":    ctx.Request().Input("map"),
						"string": ctx.Request().Input("string"),
					})
				})

				form := neturl.Values{
					"id":     {"4"},
					"map":    {"{\"a\":\"b\"}"},
					"string": {"string 1", "string 2"},
				}

				req, err = http.NewRequest(method, url, strings.NewReader(form.Encode()))
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"4\",\"map\":\"{\\\"a\\\":\\\"b\\\"}\",\"string\":\"string 1,string 2\"}",
		},
		{
			name:   "Input - from json, then Bind",
			method: "POST",
			url:    "/input/json/1?id=2",
			setup: func(method, url string) error {
				gin.Post("/input/json/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					id := ctx.Request().Input("id")
					var data struct {
						Name string `form:"name" json:"name"`
					}
					_ = ctx.Request().Bind(&data)
					return ctx.Response().Success().Json(contractshttp.Json{
						"id":   id,
						"name": data.Name,
					})
				})

				payload := strings.NewReader(`{
					"name": "Goravel"
				}`)
				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"2\",\"name\":\"Goravel\"}",
		},
		{
			name:   "Input - from form, then Bind",
			method: "POST",
			url:    "/input/form/1?id=2",
			setup: func(method, url string) error {
				gin.Post("/input/form/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					id := ctx.Request().Input("id")
					var data struct {
						Name string `form:"name" json:"name"`
					}
					_ = ctx.Request().Bind(&data)
					return ctx.Response().Success().Json(contractshttp.Json{
						"id":   id,
						"name": data.Name,
					})
				})

				payload := &bytes.Buffer{}
				writer := multipart.NewWriter(payload)
				if err := writer.WriteField("name", "Goravel"); err != nil {
					return err
				}
				if err := writer.Close(); err != nil {
					return err
				}

				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", writer.FormDataContentType())

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"2\",\"name\":\"Goravel\"}",
		},
		{
			name:   "Input - from query",
			method: "POST",
			url:    "/input3/1?id=2",
			setup: func(method, url string) error {
				gin.Post("/input3/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"id": ctx.Request().Input("id"),
					})
				})

				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"2\"}",
		},
		{
			name:   "Input - from route",
			method: "POST",
			url:    "/input4/1",
			setup: func(method, url string) error {
				gin.Post("/input4/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"id": ctx.Request().Input("id"),
					})
				})

				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name:   "Input - empty",
			method: "POST",
			url:    "/input5/1",
			setup: func(method, url string) error {
				gin.Post("/input5/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"id1": ctx.Request().Input("id1"),
					})
				})

				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id1\":\"\"}",
		},
		{
			name:   "Input - default",
			method: "POST",
			url:    "/input6/1",
			setup: func(method, url string) error {
				gin.Post("/input6/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"id1": ctx.Request().Input("id1", "2"),
					})
				})

				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id1\":\"2\"}",
		},
		{
			name:   "Input with point - application/json",
			method: "POST",
			url:    "/input/json/point/1?id=2",
			setup: func(method, url string) error {
				gin.Post("/input/json/point/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"id":      ctx.Request().Input("id.a"),
						"string0": ctx.Request().Input("string.0"),
						"string":  ctx.Request().Input("string"),
					})
				})

				payload := strings.NewReader(`{
					"id": {"a": {"b": "c"}},
					"string": ["string 0", "string 1"]
				}`)
				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"{\\\"b\\\":\\\"c\\\"}\",\"string\":\"string 0,string 1\",\"string0\":\"string 0\"}",
		},
		{
			name:   "Input with point - multipart/form-data",
			method: "POST",
			url:    "/input/form/point/1?id=2",
			setup: func(method, url string) error {
				gin.Post("/input/form/point/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"name":    ctx.Request().Input("name"),
						"string0": ctx.Request().Input("string.0"),
						"string":  ctx.Request().Input("string"),
					})
				})

				payload := &bytes.Buffer{}
				writer := multipart.NewWriter(payload)
				if err := writer.WriteField("name", "goravel"); err != nil {
					return err
				}
				if err := writer.WriteField("string[]", "string 0"); err != nil {
					return err
				}
				if err := writer.WriteField("string[]", "string 1"); err != nil {
					return err
				}
				if err := writer.Close(); err != nil {
					return err
				}

				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", writer.FormDataContentType())

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"name\":\"goravel\",\"string\":\"string 0,string 1\",\"string0\":\"string 0\"}",
		},
		{
			name:   "Input with point - application/x-www-form-urlencoded",
			method: "POST",
			url:    "/input/url/point/1?id=2",
			setup: func(method, url string) error {
				gin.Post("/input/url/point/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"id":      ctx.Request().Input("id"),
						"string0": ctx.Request().Input("string.0"),
						"string":  ctx.Request().Input("string"),
					})
				})

				form := neturl.Values{
					"id":     {"4"},
					"string": {"string 0", "string 1"},
				}

				req, err = http.NewRequest(method, url, strings.NewReader(form.Encode()))
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"4\",\"string\":\"string 0,string 1\",\"string0\":\"string 0\"}",
		},
		{
			name:   "InputArray - default",
			method: "POST",
			url:    "/input-array/default/1?id=2",
			setup: func(method, url string) error {
				gin.Post("/input-array/default/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"name": ctx.Request().InputArray("name", []string{"a", "b"}),
					})
				})

				payload := strings.NewReader(`{
					"id": ["id 0", "id 1"]
				}`)
				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}
				req.Header.Set("Content-Type", "application/json")
				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"name\":[\"a\",\"b\"]}",
		},
		{
			name:   "InputArray - empty",
			method: "POST",
			url:    "/input-array/default/1?id=2",
			setup: func(method, url string) error {
				gin.Post("/input-array/default/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"name": ctx.Request().InputArray("name"),
					})
				})

				payload := strings.NewReader(`{
					"id": ["id 0", "id 1"]
				}`)
				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"name\":[]}",
		},
		{
			name:   "InputArray of application/json",
			method: "POST",
			url:    "/input-array/json/1?id=2",
			setup: func(method, url string) error {
				gin.Post("/input-array/json/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"id": ctx.Request().InputArray("id"),
					})
				})

				payload := strings.NewReader(`{
					"id": ["id 0", "id 1"]
				}`)
				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":[\"id 0\",\"id 1\"]}",
		},
		{
			name:   "InputArray of multipart/form-data",
			method: "POST",
			url:    "/input-array/form/1?id=2",
			setup: func(method, url string) error {
				gin.Post("/input-array/form/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"InputArray":  ctx.Request().InputArray("arr[]"),
						"InputArray1": ctx.Request().InputArray("arr"),
					})
				})

				payload := &bytes.Buffer{}
				writer := multipart.NewWriter(payload)
				if err := writer.WriteField("arr[]", "arr 1"); err != nil {
					return err
				}
				if err := writer.WriteField("arr[]", "arr 2"); err != nil {
					return err
				}
				if err := writer.Close(); err != nil {
					return err
				}

				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", writer.FormDataContentType())

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"InputArray\":[\"arr 1\",\"arr 2\"],\"InputArray1\":[\"arr 1\",\"arr 2\"]}",
		},
		{
			name:   "InputArray of application/x-www-form-urlencoded",
			method: "POST",
			url:    "/input-array/url/1?id=2",
			setup: func(method, url string) error {
				gin.Post("/input-array/url/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"string":  ctx.Request().InputArray("string[]"),
						"string1": ctx.Request().InputArray("string"),
					})
				})

				form := neturl.Values{
					"string[]": {"string 0", "string 1"},
				}

				req, err = http.NewRequest(method, url, strings.NewReader(form.Encode()))
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"string\":[\"string 0\",\"string 1\"],\"string1\":[\"string 0\",\"string 1\"]}",
		},
		{
			name:   "InputMap - default",
			method: "POST",
			url:    "/input-map/default/1?id=2",
			setup: func(method, url string) error {
				gin.Post("/input-map/default/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"name": ctx.Request().InputMap("name", map[string]string{"a": "b"}),
					})
				})

				payload := strings.NewReader(`{
					"id": {"a": "3", "b": "4"}
				}`)
				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}
				req.Header.Set("Content-Type", "application/json")
				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"name\":{\"a\":\"b\"}}",
		},
		{
			name:   "InputMap - empty",
			method: "POST",
			url:    "/input-map/default/1?id=2",
			setup: func(method, url string) error {
				gin.Post("/input-map/default/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"name": ctx.Request().InputMap("name"),
					})
				})

				payload := strings.NewReader(`{
					"id": {"a": "3", "b": "4"}
				}`)
				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"name\":{}}",
		},
		{
			name:   "InputMap - application/json",
			method: "POST",
			url:    "/input-map/json/1?id=2",
			setup: func(method, url string) error {
				gin.Post("/input-map/json/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"id": ctx.Request().InputMap("id"),
					})
				})
				payload := strings.NewReader(`{
					"id": {"a": "3", "b": "4"}
				}`)
				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}
				req.Header.Set("Content-Type", "application/json")
				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":{\"a\":\"3\",\"b\":\"4\"}}",
		},
		{
			name:   "InputMap - multipart/form-data",
			method: "POST",
			url:    "/input-map/form/1?id=2",
			setup: func(method, url string) error {
				gin.Post("/input-map/form/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"id": ctx.Request().InputMap("id"),
					})
				})

				payload := &bytes.Buffer{}
				writer := multipart.NewWriter(payload)
				if err := writer.WriteField("id", "{\"a\":\"3\",\"b\":\"4\"}"); err != nil {
					return err
				}
				if err := writer.Close(); err != nil {
					return err
				}

				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", writer.FormDataContentType())

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":{\"a\":\"3\",\"b\":\"4\"}}",
		},
		{
			name:   "InputMap - application/x-www-form-urlencoded",
			method: "POST",
			url:    "/input-map/url/1?id=2",
			setup: func(method, url string) error {
				gin.Post("/input-map/url/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"id": ctx.Request().InputMap("id"),
					})
				})

				form := neturl.Values{
					"id": {"{\"a\":\"3\",\"b\":\"4\"}"},
				}

				req, err = http.NewRequest(method, url, strings.NewReader(form.Encode()))
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":{\"a\":\"3\",\"b\":\"4\"}}",
		},
		{
			name:   "InputInt",
			method: "POST",
			url:    "/input-int/1",
			setup: func(method, url string) error {
				gin.Post("/input-int/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"id": ctx.Request().InputInt("id"),
					})
				})

				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":1}",
		},
		{
			name:   "InputInt64",
			method: "POST",
			url:    "/input-int64/1",
			setup: func(method, url string) error {
				gin.Post("/input-int64/{id}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"id": ctx.Request().InputInt64("id"),
					})
				})

				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":1}",
		},
		{
			name:   "InputBool",
			method: "POST",
			url:    "/input-bool/1/true/on/yes/a",
			setup: func(method, url string) error {
				gin.Post("/input-bool/{id1}/{id2}/{id3}/{id4}/{id5}", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"id1": ctx.Request().InputBool("id1"),
						"id2": ctx.Request().InputBool("id2"),
						"id3": ctx.Request().InputBool("id3"),
						"id4": ctx.Request().InputBool("id4"),
						"id5": ctx.Request().InputBool("id5"),
					})
				})

				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id1\":true,\"id2\":true,\"id3\":true,\"id4\":true,\"id5\":false}",
		},
		{
			name:   "Bind",
			method: "POST",
			url:    "/bind",
			setup: func(method, url string) error {
				gin.Post("/bind", func(ctx contractshttp.Context) contractshttp.Response {
					type Test struct {
						Name string
					}
					var test Test
					_ = ctx.Request().Bind(&test)
					return ctx.Response().Success().Json(contractshttp.Json{
						"name": test.Name,
					})
				})

				payload := strings.NewReader(`{
					"Name": "Goravel"
				}`)
				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"name\":\"Goravel\"}",
		},
		{
			name:   "Bind, then Input",
			method: "POST",
			url:    "/bind",
			setup: func(method, url string) error {
				gin.Post("/bind", func(ctx contractshttp.Context) contractshttp.Response {
					type Test struct {
						Name string
					}
					var test Test
					_ = ctx.Request().Bind(&test)
					return ctx.Response().Success().Json(contractshttp.Json{
						"name":  test.Name,
						"name1": ctx.Request().Input("Name"),
					})
				})

				payload := strings.NewReader(`{
					"Name": "Goravel"
				}`)
				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"name\":\"Goravel\",\"name1\":\"Goravel\"}",
		},
		{
			name:   "Query",
			method: "GET",
			url:    "/query?string=Goravel&int=1&int64=2&bool1=1&bool2=true&bool3=on&bool4=yes&bool5=0&error=a",
			setup: func(method, url string) error {
				gin.Get("/query", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"string":        ctx.Request().Query("string", ""),
						"int":           ctx.Request().QueryInt("int", 11),
						"int_default":   ctx.Request().QueryInt("int_default", 11),
						"int64":         ctx.Request().QueryInt64("int64", 22),
						"int64_default": ctx.Request().QueryInt64("int64_default", 22),
						"bool1":         ctx.Request().QueryBool("bool1"),
						"bool2":         ctx.Request().QueryBool("bool2"),
						"bool3":         ctx.Request().QueryBool("bool3"),
						"bool4":         ctx.Request().QueryBool("bool4"),
						"bool5":         ctx.Request().QueryBool("bool5"),
						"error":         ctx.Request().QueryInt("error", 33),
					})
				})

				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"bool1\":true,\"bool2\":true,\"bool3\":true,\"bool4\":true,\"bool5\":false,\"error\":0,\"int\":1,\"int64\":2,\"int64_default\":22,\"int_default\":11,\"string\":\"Goravel\"}",
		},
		{
			name:   "QueryArray",
			method: "GET",
			url:    "/query-array?name=Goravel&name=Goravel1",
			setup: func(method, url string) error {
				gin.Get("/query-array", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"name": ctx.Request().QueryArray("name"),
					})
				})

				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"name\":[\"Goravel\",\"Goravel1\"]}",
		},
		{
			name:   "QueryMap",
			method: "GET",
			url:    "/query-map?name[a]=Goravel&name[b]=Goravel1",
			setup: func(method, url string) error {
				gin.Get("/query-map", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"name": ctx.Request().QueryMap("name"),
					})
				})

				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"name\":{\"a\":\"Goravel\",\"b\":\"Goravel1\"}}",
		},
		{
			name:   "Queries",
			method: "GET",
			url:    "/queries?string=Goravel&int=1&int64=2&bool1=1&bool2=true&bool3=on&bool4=yes&bool5=0&error=a",
			setup: func(method, url string) error {
				gin.Get("/queries", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"all": ctx.Request().All(),
					})
				})

				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"all\":{\"bool1\":\"1\",\"bool2\":\"true\",\"bool3\":\"on\",\"bool4\":\"yes\",\"bool5\":\"0\",\"error\":\"a\",\"int\":\"1\",\"int64\":\"2\",\"string\":\"Goravel\"}}",
		},
		{
			name:   "File",
			method: "POST",
			url:    "/file",
			setup: func(method, url string) error {
				gin.Post("/file", func(ctx contractshttp.Context) contractshttp.Response {
					mockConfig.On("GetString", "app.name").Return("goravel").Once()
					mockConfig.On("GetString", "filesystems.default").Return("local").Once()
					frameworkfilesystem.ConfigFacade = mockConfig

					mockStorage := &filesystemmocks.Storage{}
					mockDriver := &filesystemmocks.Driver{}
					mockStorage.On("Disk", "local").Return(mockDriver).Once()
					frameworkfilesystem.StorageFacade = mockStorage

					fileInfo, err := ctx.Request().File("file")

					mockDriver.On("PutFile", "test", fileInfo).Return("test/README.md", nil).Once()
					mockStorage.On("Exists", "test/README.md").Return(true).Once()

					if err != nil {
						return ctx.Response().Success().String("get file error")
					}
					filePath, err := fileInfo.Store("test")
					if err != nil {
						return ctx.Response().Success().String("store file error: " + err.Error())
					}

					extension, err := fileInfo.Extension()
					if err != nil {
						return ctx.Response().Success().String("get file extension error: " + err.Error())
					}

					return ctx.Response().Success().Json(contractshttp.Json{
						"exist":              mockStorage.Exists(filePath),
						"hash_name_length":   len(fileInfo.HashName()),
						"hash_name_length1":  len(fileInfo.HashName("test")),
						"file_path_length":   len(filePath),
						"extension":          extension,
						"original_name":      fileInfo.GetClientOriginalName(),
						"original_extension": fileInfo.GetClientOriginalExtension(),
					})
				})

				payload := &bytes.Buffer{}
				writer := multipart.NewWriter(payload)
				readme, err := os.Open("./README.md")
				if err != nil {
					return err
				}
				defer readme.Close()
				part1, err := writer.CreateFormFile("file", filepath.Base("./README.md"))
				if err != nil {
					return err
				}

				if _, err = io.Copy(part1, readme); err != nil {
					return err
				}

				if err := writer.Close(); err != nil {
					return err
				}

				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", writer.FormDataContentType())

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"exist\":true,\"extension\":\"txt\",\"file_path_length\":14,\"hash_name_length\":44,\"hash_name_length1\":49,\"original_extension\":\"md\",\"original_name\":\"README.md\"}",
		},
		{
			name:   "GET with params and validator, validate pass",
			method: "GET",
			url:    "/validator/validate/success/abc?name=Goravel",
			setup: func(method, url string) error {
				gin.Get("/validator/validate/success/{uuid}", func(ctx contractshttp.Context) contractshttp.Response {
					mockValication := &validationmocks.Validation{}
					mockValication.On("Rules").Return([]validation.Rule{}).Once()
					ValidationFacade = mockValication

					validator, err := ctx.Request().Validate(map[string]string{
						"uuid": "min_len:2",
						"name": "required",
					})
					if err != nil {
						return ctx.Response().String(400, "Validate error: "+err.Error())
					}
					if validator.Fails() {
						return ctx.Response().String(400, fmt.Sprintf("Validate fail: %+v", validator.Errors().All()))
					}

					type Test struct {
						Uuid string `form:"uuid" json:"uuid"`
						Name string `form:"name" json:"name"`
					}
					var test Test
					if err := validator.Bind(&test); err != nil {
						return ctx.Response().String(400, "Validate bind error: "+err.Error())
					}

					return ctx.Response().Success().Json(contractshttp.Json{
						"uuid": test.Uuid,
						"name": test.Name,
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
			expectBody: "{\"name\":\"Goravel\",\"uuid\":\"abc\"}",
		},
		{
			name:   "GET with params and validator, validate fail",
			method: "GET",
			url:    "/validator/validate/fail/abc?name=Goravel",
			setup: func(method, url string) error {
				gin.Get("/validator/validate/fail/{uuid}", func(ctx contractshttp.Context) contractshttp.Response {
					mockValication := &validationmocks.Validation{}
					mockValication.On("Rules").Return([]validation.Rule{}).Once()
					ValidationFacade = mockValication

					validator, err := ctx.Request().Validate(map[string]string{
						"uuid": "min_len:4",
						"name": "required",
					})
					if err != nil {
						return ctx.Response().String(400, "Validate error: "+err.Error())
					}
					if validator.Fails() {
						return ctx.Response().String(400, fmt.Sprintf("Validate fail: %+v", validator.Errors().All()))
					}

					return nil
				})
				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusBadRequest,
			expectBody: "Validate fail: map[uuid:map[min_len:uuid min length is 4]]",
		},
		{
			name:   "GET with validator and validate pass",
			method: "GET",
			url:    "/validator/validate/success?name=Goravel",
			setup: func(method, url string) error {
				gin.Get("/validator/validate/success", func(ctx contractshttp.Context) contractshttp.Response {
					mockValication := &validationmocks.Validation{}
					mockValication.On("Rules").Return([]validation.Rule{}).Once()
					ValidationFacade = mockValication

					validator, err := ctx.Request().Validate(map[string]string{
						"name": "required",
					})
					if err != nil {
						return ctx.Response().String(400, "Validate error: "+err.Error())
					}
					if validator.Fails() {
						return ctx.Response().String(400, fmt.Sprintf("Validate fail: %+v", validator.Errors().All()))
					}

					type Test struct {
						Name string `form:"name" json:"name"`
					}
					var test Test
					if err := validator.Bind(&test); err != nil {
						return ctx.Response().String(400, "Validate bind error: "+err.Error())
					}

					return ctx.Response().Success().Json(contractshttp.Json{
						"name": test.Name,
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
			expectBody: "{\"name\":\"Goravel\"}",
		},
		{
			name:   "GET with validator but validate fail",
			method: "GET",
			url:    "/validator/validate/fail?name=Goravel",
			setup: func(method, url string) error {
				gin.Get("/validator/validate/fail", func(ctx contractshttp.Context) contractshttp.Response {
					mockValication := &validationmocks.Validation{}
					mockValication.On("Rules").Return([]validation.Rule{}).Once()
					ValidationFacade = mockValication

					validator, err := ctx.Request().Validate(map[string]string{
						"name1": "required",
					})
					if err != nil {
						return ctx.Response().String(http.StatusBadRequest, "Validate error: "+err.Error())
					}
					if validator.Fails() {
						return ctx.Response().String(http.StatusBadRequest, fmt.Sprintf("Validate fail: %+v", validator.Errors().All()))
					}

					return nil
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusBadRequest,
			expectBody: "Validate fail: map[name1:map[required:name1 is required to not be empty]]",
		},
		{
			name:   "GET with validator and validate request pass",
			method: "GET",
			url:    "/validator/validate-request/success?name=Goravel",
			setup: func(method, url string) error {
				gin.Get("/validator/validate-request/success", func(ctx contractshttp.Context) contractshttp.Response {
					mockValication := &validationmocks.Validation{}
					mockValication.On("Rules").Return([]validation.Rule{}).Once()
					ValidationFacade = mockValication

					var createUser CreateUser
					validateErrors, err := ctx.Request().ValidateRequest(&createUser)
					if err != nil {
						return ctx.Response().String(http.StatusBadRequest, "Validate error: "+err.Error())
					}
					if validateErrors != nil {
						return ctx.Response().String(http.StatusBadRequest, fmt.Sprintf("Validate fail: %+v", validateErrors.All()))
					}

					return ctx.Response().Success().Json(contractshttp.Json{
						"name": createUser.Name,
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
			expectBody: "{\"name\":\"Goravel1\"}",
		},
		{
			name:   "GET with validator but validate request fail",
			method: "GET",
			url:    "/validator/validate-request/fail?name1=Goravel",
			setup: func(method, url string) error {
				gin.Get("/validator/validate-request/fail", func(ctx contractshttp.Context) contractshttp.Response {
					mockValication := &validationmocks.Validation{}
					mockValication.On("Rules").Return([]validation.Rule{}).Once()
					ValidationFacade = mockValication

					var createUser CreateUser
					validateErrors, err := ctx.Request().ValidateRequest(&createUser)
					if err != nil {
						return ctx.Response().String(http.StatusBadRequest, "Validate error: "+err.Error())
					}
					if validateErrors != nil {
						return ctx.Response().String(http.StatusBadRequest, fmt.Sprintf("Validate fail: %+v", validateErrors.All()))
					}

					return ctx.Response().Success().Json(contractshttp.Json{
						"name": createUser.Name,
					})
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusBadRequest,
			expectBody: "Validate fail: map[name:map[required:name is required to not be empty]]",
		},
		{
			name:   "GET with validator use filter and validate request pass",
			method: "GET",
			url:    "/validator/filter/success?name= Goravel ",
			setup: func(method, url string) error {
				gin.Get("/validator/filter/success", func(ctx contractshttp.Context) contractshttp.Response {
					mockValidation := &validationmocks.Validation{}
					mockValidation.On("Rules").Return([]validation.Rule{}).Once()
					ValidationFacade = mockValidation

					var createUser CreateUser
					validateErrors, err := ctx.Request().ValidateRequest(&createUser)
					if err != nil {
						return ctx.Response().String(http.StatusBadRequest, "Validate error: "+err.Error())
					}
					if validateErrors != nil {
						return ctx.Response().String(http.StatusBadRequest, fmt.Sprintf("Validate fail: %+v", validateErrors.All()))
					}

					return ctx.Response().Success().Json(contractshttp.Json{
						"name": createUser.Name,
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
			expectBody: "{\"name\":\"Goravel 1\"}",
		},
		{
			name:   "POST with validator and validate pass",
			method: "POST",
			url:    "/validator/validate/success",
			setup: func(method, url string) error {
				gin.Post("/validator/validate/success", func(ctx contractshttp.Context) contractshttp.Response {
					mockValication := &validationmocks.Validation{}
					mockValication.On("Rules").Return([]validation.Rule{}).Once()
					ValidationFacade = mockValication

					validator, err := ctx.Request().Validate(map[string]string{
						"name": "required",
					})
					if err != nil {
						return ctx.Response().String(400, "Validate error: "+err.Error())
					}
					if validator.Fails() {
						return ctx.Response().String(400, fmt.Sprintf("Validate fail: %+v", validator.Errors().All()))
					}

					type Test struct {
						Name string `form:"name" json:"name"`
					}
					var test Test
					if err := validator.Bind(&test); err != nil {
						return ctx.Response().String(400, "Validate bind error: "+err.Error())
					}

					return ctx.Response().Success().Json(contractshttp.Json{
						"name": test.Name,
					})
				})

				payload := strings.NewReader(`{
					"name": "Goravel"
				}`)
				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"name\":\"Goravel\"}",
		},
		{
			name:   "POST with validator and validate fail",
			method: "POST",
			url:    "/validator/validate/fail",
			setup: func(method, url string) error {
				gin.Post("/validator/validate/fail", func(ctx contractshttp.Context) contractshttp.Response {
					mockValication := &validationmocks.Validation{}
					mockValication.On("Rules").Return([]validation.Rule{}).Once()
					ValidationFacade = mockValication

					validator, err := ctx.Request().Validate(map[string]string{
						"name1": "required",
					})
					if err != nil {
						return ctx.Response().String(400, "Validate error: "+err.Error())
					}
					if validator.Fails() {
						return ctx.Response().String(400, fmt.Sprintf("Validate fail: %+v", validator.Errors().All()))
					}

					return ctx.Response().Success().Json(contractshttp.Json{
						"name": "",
					})
				})
				payload := strings.NewReader(`{
					"name": "Goravel"
				}`)
				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusBadRequest,
			expectBody: "Validate fail: map[name1:map[required:name1 is required to not be empty]]",
		},
		{
			name:   "POST with validator and validate request pass",
			method: "POST",
			url:    "/validator/validate-request/success",
			setup: func(method, url string) error {
				gin.Post("/validator/validate-request/success", func(ctx contractshttp.Context) contractshttp.Response {
					mockValication := &validationmocks.Validation{}
					mockValication.On("Rules").Return([]validation.Rule{}).Once()
					ValidationFacade = mockValication

					var createUser CreateUser
					validateErrors, err := ctx.Request().ValidateRequest(&createUser)
					if err != nil {
						return ctx.Response().String(http.StatusBadRequest, "Validate error: "+err.Error())
					}
					if validateErrors != nil {
						return ctx.Response().String(http.StatusBadRequest, fmt.Sprintf("Validate fail: %+v", validateErrors.All()))
					}

					return ctx.Response().Success().Json(contractshttp.Json{
						"name": createUser.Name,
					})
				})

				payload := strings.NewReader(`{
					"name": "Goravel"
				}`)
				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"name\":\"Goravel1\"}",
		},
		{
			name:   "POST with validator and validate request fail",
			method: "POST",
			url:    "/validator/validate-request/fail",
			setup: func(method, url string) error {
				gin.Post("/validator/validate-request/fail", func(ctx contractshttp.Context) contractshttp.Response {
					mockValication := &validationmocks.Validation{}
					mockValication.On("Rules").Return([]validation.Rule{}).Once()
					ValidationFacade = mockValication

					var createUser CreateUser
					validateErrors, err := ctx.Request().ValidateRequest(&createUser)
					if err != nil {
						return ctx.Response().String(http.StatusBadRequest, "Validate error: "+err.Error())
					}
					if validateErrors != nil {
						return ctx.Response().String(http.StatusBadRequest, fmt.Sprintf("Validate fail: %+v", validateErrors.All()))
					}

					return ctx.Response().Success().Json(contractshttp.Json{
						"name": createUser.Name,
					})
				})

				payload := strings.NewReader(`{
					"name1": "Goravel"
				}`)
				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusBadRequest,
			expectBody: "Validate fail: map[name:map[required:name is required to not be empty]]",
		},
		{name: "POST with validator use filter and validate request pass",
			method: "POST",
			url:    "/validator/filter/success",
			setup: func(method, url string) error {
				gin.Post("/validator/filter/success", func(ctx contractshttp.Context) contractshttp.Response {
					mockValidation := &validationmocks.Validation{}
					mockValidation.On("Rules").Return([]validation.Rule{}).Once()
					ValidationFacade = mockValidation

					var createUser CreateUser
					validateErrors, err := ctx.Request().ValidateRequest(&createUser)
					if err != nil {
						return ctx.Response().String(http.StatusBadRequest, "Validate error: "+err.Error())
					}
					if validateErrors != nil {
						return ctx.Response().String(http.StatusBadRequest, fmt.Sprintf("Validate fail: %+v", validateErrors.All()))
					}

					return ctx.Response().Success().Json(contractshttp.Json{
						"name": createUser.Name,
					})
				})

				payload := strings.NewReader(`{
					"name": " Goravel "
				}`)
				req, _ = http.NewRequest(method, url, payload)
				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"name\":\"Goravel 1\"}",
		},
		{
			name:   "POST with validator and validate request unauthorize",
			method: "POST",
			url:    "/validator/validate-request/unauthorize",
			setup: func(method, url string) error {
				gin.Post("/validator/validate-request/unauthorize", func(ctx contractshttp.Context) contractshttp.Response {
					var unauthorize Unauthorize
					validateErrors, err := ctx.Request().ValidateRequest(&unauthorize)
					if err != nil {
						return ctx.Response().String(http.StatusBadRequest, "Validate error: "+err.Error())
					}
					if validateErrors != nil {
						return ctx.Response().String(http.StatusBadRequest, fmt.Sprintf("Validate fail: %+v", validateErrors.All()))
					}

					return ctx.Response().Success().Json(contractshttp.Json{
						"name": unauthorize.Name,
					})
				})
				payload := strings.NewReader(`{
					"name": "Goravel"
				}`)
				req, err = http.NewRequest(method, url, payload)
				if err != nil {
					return err
				}

				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusBadRequest,
			expectBody: "Validate error: error",
		},
		{
			name:   "Cookie",
			method: "GET",
			url:    "/cookie",
			setup: func(method, url string) error {
				gin.Get("/cookie", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"goravel": ctx.Request().Cookie("goravel"),
					})
				})

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
			expectCode: http.StatusOK,
			expectBody: "{\"goravel\":\"goravel\"}",
		},
		{
			name:   "Cookie - default value",
			method: "GET",
			url:    "/cookie/default",
			setup: func(method, url string) error {
				gin.Get("/cookie/default", func(ctx contractshttp.Context) contractshttp.Response {
					return ctx.Response().Success().Json(contractshttp.Json{
						"goravel": ctx.Request().Cookie("goravel", "default value"),
					})
				})

				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"goravel\":\"default value\"}",
		},
		{
			name:   "Session",
			method: "GET",
			url:    "/session",
			setup: func(method, url string) error {
				gin.Get("/session", func(ctx contractshttp.Context) contractshttp.Response {
					ctx.Request().SetSession(session.NewSession("goravel_session", nil, foundationjson.NewJson()))

					return ctx.Response().Success().Json(contractshttp.Json{
						"message": ctx.Request().Session().GetName(),
					})
				})

				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"message\":\"goravel_session\"}",
		},
		{
			name:   "Session - not set",
			method: "GET",
			url:    "/session/not-set",
			setup: func(method, url string) error {
				gin.Get("/session/not-set", func(ctx contractshttp.Context) contractshttp.Response {
					if ctx.Request().HasSession() {
						return ctx.Response().Success().Json(contractshttp.Json{
							"message": ctx.Request().Session().GetName(),
						})
					}

					return ctx.Response().Success().Json(contractshttp.Json{
						"message": "session not set",
					})
				})

				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"message\":\"session not set\"}",
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

			assert.Equal(t, test.expectCode, w.Code)

			mockConfig.AssertExpectations(t)
		})
	}
}

func TestGetValueFromPostData(t *testing.T) {
	tests := []struct {
		name        string
		postData    map[string]any
		key         string
		expectValue any
	}{
		{
			name: "Return nil when postData is nil",
		},
		{
			name:        "Return string when postData is map[string]string",
			postData:    map[string]any{"name": "goravel"},
			key:         "name",
			expectValue: "goravel",
		},
		{
			name:        "Return map when postData is map[string]map[string]string",
			postData:    map[string]any{"name": map[string]string{"sub": "goravel"}},
			key:         "name",
			expectValue: map[string]string{"sub": "goravel"},
		},
		{
			name:        "Return slice when postData is map[string][]string",
			postData:    map[string]any{"name[]": []string{"a", "b"}},
			key:         "name[]",
			expectValue: []string{"a", "b"},
		},
		{
			name:        "Return slice when postData is map[string][]string, but key doesn't contain []",
			postData:    map[string]any{"name": []string{"a", "b"}},
			key:         "name",
			expectValue: []string{"a", "b"},
		},
		{
			name:        "Return string when postData is map[string]map[string]string and key with point",
			postData:    map[string]any{"name": map[string]string{"sub": "goravel"}},
			key:         "name.sub",
			expectValue: "goravel",
		},
		{
			name:        "Return int when postData is map[string]map[string]int and key with point",
			postData:    map[string]any{"name": map[string]int{"sub": 1}},
			key:         "name.sub",
			expectValue: 1,
		},
		{
			name:        "Return string when postData is map[string][]string and key with point",
			postData:    map[string]any{"name[]": []string{"a", "b"}},
			key:         "name[].0",
			expectValue: "a",
		},
		{
			name:        "Return string when postData is map[string][]string and key with point and index is 1",
			postData:    map[string]any{"name[]": []string{"a", "b"}},
			key:         "name[].1",
			expectValue: "b",
		},
		{
			name:        "Return string when postData is map[string][]string and key with point, but key doesn't contain []",
			postData:    map[string]any{"name[]": []string{"a", "b"}},
			key:         "name.0",
			expectValue: "a",
		},
		{
			name:        "Return string when postData is map[string][]string and key with point and index is 1, but key doesn't contain []",
			postData:    map[string]any{"name[]": []string{"a", "b"}},
			key:         "name.1",
			expectValue: "b",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contextRequest := &ContextRequest{
				postData: test.postData,
			}

			value := contextRequest.getValueFromPostData(test.key)
			assert.Equal(t, test.expectValue, value)
		})
	}
}
