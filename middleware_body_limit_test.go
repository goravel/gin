package gin

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

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
			}

			req, err := http.NewRequest(test.method, test.path, body)
			require.NoError(t, err)
			if test.chunked {
				req.ContentLength = -1
				req.TransferEncoding = []string{"chunked"}
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

	const wantLimit = 4096 << 10

	for _, test := range []struct {
		size     int
		wantCode int
	}{
		{size: wantLimit, wantCode: http.StatusOK},
		{size: wantLimit + 1, wantCode: http.StatusRequestEntityTooLarge},
	} {
		req, err := http.NewRequest(http.MethodPost, "/raw", bytes.NewReader(make([]byte, test.size)))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/octet-stream")

		w := httptest.NewRecorder()
		route.ServeHTTP(w, req)

		assert.Equal(t, test.wantCode, w.Code, "size %d", test.size)
	}
}

func TestBodyLimitMiddleware_StalledChunkedBodyReadByHandler(t *testing.T) {
	route := newBodyLimitTestRoute(t, 1)
	route.Post("/raw", func(ctx contractshttp.Context) contractshttp.Response {
		_, err := io.ReadAll(ctx.Request().Origin().Body)
		if isBodyTooLarge(err) {
			return ctx.Response().String(http.StatusBadRequest, "too large")
		}

		return ctx.Response().String(http.StatusOK, "ok")
	})

	response := stalledChunkedRequest(t, route, "/raw", "application/octet-stream")

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestBodyLimitMiddleware_StalledChunkedJson(t *testing.T) {
	route := newBodyLimitTestRoute(t, 1)
	route.Post("/input", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().String(http.StatusOK, ctx.Request().Input("name"))
	})

	response := stalledChunkedRequest(t, route, "/input", "application/json")

	assert.Equal(t, http.StatusRequestEntityTooLarge, response.StatusCode)
}

func TestBodyLimitMiddleware_ContentLengthOverLimitIsNotRead(t *testing.T) {
	route := newBodyLimitTestRoute(t, 1)
	route.Post("/input", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().String(http.StatusOK, "ok")
	})

	body := &countingReader{}
	request := httptest.NewRequest(http.MethodPost, "/input", body)
	request.ContentLength = 2048
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	route.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code)
	assert.Zero(t, body.reads, "a body already known to be over the limit must not be read")
}

// The server drains up to 256 KB of an unread body before it writes the response,
// so a declared length under that from a client that stopped sending would stall
// the 413 unless the connection is marked for close.
func TestBodyLimitMiddleware_StalledContentLengthOverLimit(t *testing.T) {
	route := newBodyLimitTestRoute(t, 1)
	route.Post("/input", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().String(http.StatusOK, "ok")
	})

	response := stalledRequest(t, route, "/input",
		"Content-Type: application/json\r\nContent-Length: 204800", strings.Repeat("a", 2048))

	assert.Equal(t, http.StatusRequestEntityTooLarge, response.StatusCode)
	assert.True(t, response.Close)
}

func newBodyLimitTestRoute(t *testing.T, limit int) *Route {
	t.Helper()

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
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "file.txt")
	require.NoError(t, err)
	_, err = part.Write(bytes.Repeat([]byte("a"), size))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	return body.String(), writer.FormDataContentType()
}

// stalledRequest sends a request whose body goes over the limit and then stops
// sending: the rest of the body never comes and the connection stays open. The
// response has to come back anyway, otherwise the server is left reading from a
// peer that says nothing and no read timeout is configured on it.
func stalledRequest(t *testing.T, route *Route, path, headers, body string) *http.Response {
	t.Helper()

	const timeout = 2 * time.Second

	server := httptest.NewServer(route)
	t.Cleanup(server.Close)

	conn, err := net.Dial("tcp", server.Listener.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	require.NoError(t, conn.SetWriteDeadline(time.Now().Add(timeout)))
	_, err = fmt.Fprintf(conn, "POST %s HTTP/1.1\r\nHost: goravel\r\n%s\r\n\r\n%s", path, headers, body)
	require.NoError(t, err)

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(timeout)))
	response, err := http.ReadResponse(bufio.NewReader(conn), nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = response.Body.Close() })

	return response
}

func stalledChunkedRequest(t *testing.T, route *Route, path, contentType string) *http.Response {
	t.Helper()

	chunk := strings.Repeat("a", 2048)

	return stalledRequest(t, route, path,
		fmt.Sprintf("Content-Type: %s\r\nTransfer-Encoding: chunked", contentType),
		fmt.Sprintf("%x\r\n%s\r\n", len(chunk), chunk))
}

// countingReader records how many times the body was read.
type countingReader struct {
	reads int
}

func (r *countingReader) Read(p []byte) (int, error) {
	r.reads++

	return 0, io.EOF
}
