package gin

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	contractshttp "github.com/goravel/framework/contracts/http"
	mocksconfig "github.com/goravel/framework/mocks/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBodyLimitMiddleware(t *testing.T) {
	route := newBodyLimitTestRoute(t, 1)

	var called bool
	route.Post("/input", func(ctx contractshttp.Context) contractshttp.Response {
		called = true
		return ctx.Response().String(http.StatusOK, ctx.Request().Input("name"))
	})
	route.Post("/raw", func(ctx contractshttp.Context) contractshttp.Response {
		called = true
		body, err := io.ReadAll(ctx.Request().Origin().Body)
		if isBodyTooLarge(err) {
			return ctx.Response().String(http.StatusBadRequest, "too large")
		}
		if err != nil {
			return ctx.Response().String(http.StatusInternalServerError, err.Error())
		}

		return ctx.Response().String(http.StatusOK, string(body))
	})
	route.Get("/get", func(ctx contractshttp.Context) contractshttp.Response {
		called = true
		return ctx.Response().String(http.StatusOK, "ok")
	})

	largeJson := `{"name":"` + strings.Repeat("a", 2048) + `"}`
	largeForm := url.Values{"name": {strings.Repeat("a", 2048)}}.Encode()
	largeMultipart, multipartContentType := multipartBody(t, 2048)

	tests := []struct {
		name        string
		method      string
		path        string
		contentType string
		body        string
		chunked     bool
		wantCode    int
		wantBody    string
		wantCalled  bool
	}{
		{
			name:        "json under the limit",
			method:      http.MethodPost,
			path:        "/input",
			contentType: "application/json",
			body:        `{"name":"goravel"}`,
			wantCode:    http.StatusOK,
			wantBody:    "goravel",
			wantCalled:  true,
		},
		{
			name:        "chunked json under the limit",
			method:      http.MethodPost,
			path:        "/input",
			contentType: "application/json",
			body:        `{"name":"goravel"}`,
			chunked:     true,
			wantCode:    http.StatusOK,
			wantBody:    "goravel",
			wantCalled:  true,
		},
		{
			name:       "request without a body",
			method:     http.MethodGet,
			path:       "/get",
			wantCode:   http.StatusOK,
			wantBody:   "ok",
			wantCalled: true,
		},
		{
			name:        "json over the limit",
			method:      http.MethodPost,
			path:        "/input",
			contentType: "application/json",
			body:        largeJson,
			wantCode:    http.StatusRequestEntityTooLarge,
			wantBody:    "Request Entity Too Large",
		},
		{
			name:        "chunked json over the limit",
			method:      http.MethodPost,
			path:        "/input",
			contentType: "application/json",
			body:        largeJson,
			chunked:     true,
			wantCode:    http.StatusRequestEntityTooLarge,
			wantBody:    "Request Entity Too Large",
		},
		{
			name:        "multipart over the limit",
			method:      http.MethodPost,
			path:        "/input",
			contentType: multipartContentType,
			body:        largeMultipart,
			wantCode:    http.StatusRequestEntityTooLarge,
			wantBody:    "Request Entity Too Large",
		},
		{
			name:        "chunked multipart over the limit",
			method:      http.MethodPost,
			path:        "/input",
			contentType: multipartContentType,
			body:        largeMultipart,
			chunked:     true,
			wantCode:    http.StatusRequestEntityTooLarge,
			wantBody:    "Request Entity Too Large",
		},
		{
			name:        "chunked form over the limit",
			method:      http.MethodPost,
			path:        "/input",
			contentType: "application/x-www-form-urlencoded",
			body:        largeForm,
			chunked:     true,
			wantCode:    http.StatusRequestEntityTooLarge,
			wantBody:    "Request Entity Too Large",
		},
		{
			name:        "json over the limit to an unmatched route",
			method:      http.MethodPost,
			path:        "/unknown",
			contentType: "application/json",
			body:        largeJson,
			wantCode:    http.StatusRequestEntityTooLarge,
			wantBody:    "Request Entity Too Large",
		},
		{
			name:        "chunked body read by the handler",
			method:      http.MethodPost,
			path:        "/raw",
			contentType: "application/octet-stream",
			body:        strings.Repeat("a", 2048),
			chunked:     true,
			wantCode:    http.StatusBadRequest,
			wantBody:    "too large",
			wantCalled:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			called = false

			var body io.Reader
			if test.body != "" {
				body = strings.NewReader(test.body)
				if test.chunked {
					// Hide the length so the request looks like a chunked one.
					body = io.MultiReader(body)
				}
			}

			req, err := http.NewRequest(test.method, test.path, body)
			require.NoError(t, err)
			if test.chunked {
				req.ContentLength = -1
			}
			if test.contentType != "" {
				req.Header.Set("Content-Type", test.contentType)
			}

			w := httptest.NewRecorder()
			route.ServeHTTP(w, req)

			assert.Equal(t, test.wantCode, w.Code)
			assert.Equal(t, test.wantBody, w.Body.String())
			assert.Equal(t, test.wantCalled, called)
		})
	}
}

func TestBodyLimitMiddleware_DefaultLimit(t *testing.T) {
	route := newBodyLimitTestRoute(t, 0)

	route.Post("/raw", func(ctx contractshttp.Context) contractshttp.Response {
		body, err := io.ReadAll(ctx.Request().Origin().Body)
		require.NoError(t, err)

		return ctx.Response().String(http.StatusOK, "%d", len(body))
	})

	for _, test := range []struct {
		size     int
		wantCode int
	}{
		{size: defaultBodyLimit, wantCode: http.StatusOK},
		{size: defaultBodyLimit + 1, wantCode: http.StatusRequestEntityTooLarge},
	} {
		req, err := http.NewRequest(http.MethodPost, "/raw", bytes.NewReader(make([]byte, test.size)))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/octet-stream")

		w := httptest.NewRecorder()
		route.ServeHTTP(w, req)

		assert.Equal(t, test.wantCode, w.Code, "size %d", test.size)
	}
}

func newBodyLimitTestRoute(t *testing.T, limit int) *Route {
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetInt("http.drivers.gin.body_limit", 4096).Return(limit).Once()
	mockConfig.EXPECT().GetBool("app.debug").Return(false).Once()
	mockConfig.EXPECT().Get("http.drivers.gin.template").Return(nil).Once()

	route := &Route{
		config: mockConfig,
		driver: "gin",
	}
	require.NoError(t, route.init(nil))

	return route
}

func multipartBody(t *testing.T, size int) (string, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "file.txt")
	require.NoError(t, err)
	_, err = part.Write(bytes.Repeat([]byte("a"), size))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	return body.String(), writer.FormDataContentType()
}
