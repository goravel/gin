package gin

import (
	"bufio"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	contractshttp "github.com/goravel/framework/contracts/http"
	mocksconfig "github.com/goravel/framework/mocks/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ContextResponseSuite struct {
	suite.Suite
	route      *Route
	mockConfig *mocksconfig.Config
}

func TestContextResponseSuite(t *testing.T) {
	suite.Run(t, new(ContextResponseSuite))
}

func (s *ContextResponseSuite) SetupTest() {
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockConfig.EXPECT().GetInt("http.drivers.gin.body_limit", 4096).Return(4096).Once()
	s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	s.mockConfig.EXPECT().Get("http.drivers.gin.template").Return(nil).Once()

	s.route = &Route{
		config: s.mockConfig,
		driver: "gin",
	}
	err := s.route.init(nil)
	s.Require().Nil(err)
}

func (s *ContextResponseSuite) TestCookie() {
	cookieName := "goravel"
	s.route.Get("/cookie", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Cookie(contractshttp.Cookie{
			Name:  cookieName,
			Value: "Goravel",
		}).String(http.StatusOK, "Goravel")
	})

	code, body, _, cookies := s.request("GET", "/cookie", nil)

	s.Equal("Goravel", body)
	s.Equal(http.StatusOK, code)
	exist := false
	for _, cookie := range cookies {
		if cookie.Name == cookieName {
			exist = true
			s.Equal("Goravel", cookie.Value)
		}
	}
	s.True(exist)
}

func (s *ContextResponseSuite) TestCookie_SameNameReplaced() {
	cookieName := "goravel"
	s.route.Get("/cookie-replace", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Cookie(contractshttp.Cookie{
			Name:  cookieName,
			Value: "stale",
		}).Cookie(contractshttp.Cookie{
			Name:  "other",
			Value: "kept",
		}).Cookie(contractshttp.Cookie{
			Name:  cookieName,
			Value: "rotated",
		}).String(http.StatusOK, "Goravel")
	})

	code, _, _, cookies := s.request("GET", "/cookie-replace", nil)

	s.Equal(http.StatusOK, code)

	var values []string
	otherKept := false
	for _, cookie := range cookies {
		if cookie.Name == cookieName {
			values = append(values, cookie.Value)
		}
		if cookie.Name == "other" {
			otherKept = true
		}
	}
	s.Equal([]string{"rotated"}, values)
	s.True(otherKept)
}

func (s *ContextResponseSuite) TestWithoutCookie_AfterCookie() {
	cookieName := "goravel"
	s.route.Get("/cookie-clear", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Cookie(contractshttp.Cookie{
			Name:  cookieName,
			Value: "Goravel",
		}).WithoutCookie(cookieName).String(http.StatusOK, "Goravel")
	})

	code, _, _, cookies := s.request("GET", "/cookie-clear", nil)

	s.Equal(http.StatusOK, code)

	var matched []*http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == cookieName {
			matched = append(matched, cookie)
		}
	}
	s.Require().Len(matched, 1)
	s.Empty(matched[0].Value)
	s.Equal(-1, matched[0].MaxAge)
}

func (s *ContextResponseSuite) TestData() {
	s.route.Get("/data", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Data(http.StatusOK, "text/html; charset=utf-8", []byte("<b>Goravel</b>"))
	})

	code, body, _, _ := s.request("GET", "/data", nil)

	s.Equal("<b>Goravel</b>", body)
	s.Equal(http.StatusOK, code)
}

func (s *ContextResponseSuite) TestDownload() {
	s.route.Get("/download", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Download("./test.txt", "README.md")
	})

	code, body, _, _ := s.request("GET", "/download", nil)

	s.Equal("Goravel", body)
	s.Equal(http.StatusOK, code)
}

func (s *ContextResponseSuite) TestFile() {
	s.route.Get("/file", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().File("./test.txt")
	})

	code, body, _, _ := s.request("GET", "/file", nil)

	s.Equal("Goravel", body)
	s.Equal(http.StatusOK, code)
}

func (s *ContextResponseSuite) TestHeader() {
	s.route.Get("/header", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Header("Hello", "Goravel").String(http.StatusOK, "Goravel")
	})

	code, body, header, _ := s.request("GET", "/header", nil)

	s.Equal("Goravel", body)
	s.Equal(http.StatusOK, code)
	s.Equal("Goravel", strings.Join(header.Values("Hello"), ""))
}

func (s *ContextResponseSuite) TestJson() {
	s.route.Get("/json", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Json(http.StatusOK, contractshttp.Json{
			"id": "1",
		})
	})

	code, body, _, _ := s.request("GET", "/json", nil)

	s.Equal("{\"id\":\"1\"}", body)
	s.Equal(http.StatusOK, code)
}

func (s *ContextResponseSuite) TestNoContent() {
	s.route.Get("/no-content", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().NoContent()
	})

	code, body, _, _ := s.request("GET", "/no-content", nil)

	s.Empty(body)
	s.Equal(http.StatusNoContent, code)
}

func (s *ContextResponseSuite) TestNoContent_WithCode() {
	s.route.Get("/no-content-with-code", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().NoContent(http.StatusAccepted)
	})

	code, body, _, _ := s.request("GET", "/no-content-with-code", nil)

	s.Empty(body)
	s.Equal(http.StatusAccepted, code)
}

func (s *ContextResponseSuite) TestOrigin() {
	s.mockConfig.EXPECT().GetInt("http.drivers.gin.body_limit", 4096).Return(4096).Once()
	s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	s.mockConfig.EXPECT().Get("http.drivers.gin.template").Return(nil).Once()

	s.route.GlobalMiddleware(&contextResponseTestMiddleware{s: s})
	s.route.Get("/origin", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().String(http.StatusOK, "Goravel")
	})

	code, body, _, _ := s.request("GET", "/origin", nil)

	s.Equal("Goravel", body)
	s.Equal(http.StatusOK, code)
}

func (s *ContextResponseSuite) TestRedirect() {
	s.route.Get("/redirect", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Redirect(http.StatusMovedPermanently, "https://goravel.dev")
	})

	code, body, _, _ := s.request("GET", "/redirect", nil)

	s.Equal("<a href=\"https://goravel.dev\">Moved Permanently</a>.\n\n", body)
	s.Equal(http.StatusMovedPermanently, code)
}

func (s *ContextResponseSuite) TestStream() {
	s.route.Get("/stream", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Stream(http.StatusCreated, func(w contractshttp.StreamWriter) error {
			b := []string{"a", "b", "c"}
			for _, a := range b {
				if _, err := w.Write([]byte(a + "\n")); err != nil {
					return err
				}

				if err := w.Flush(); err != nil {
					return err
				}
			}

			return nil
		})
	})

	code, body, _, _ := s.request("GET", "/stream", nil)

	scanner := bufio.NewScanner(strings.NewReader(body))
	var output []string
	for scanner.Scan() {
		output = append(output, scanner.Text())
	}

	s.Equal([]string{"a", "b", "c"}, output)
	s.Equal(http.StatusCreated, code)
}

func (s *ContextResponseSuite) TestString() {
	s.route.Get("/string", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().String(http.StatusCreated, "Goravel")
	})

	code, body, _, _ := s.request("GET", "/string", nil)

	s.Equal("Goravel", body)
	s.Equal(http.StatusCreated, code)
}

func (s *ContextResponseSuite) TestSuccess_Data() {
	s.route.Get("/data", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Success().Data("text/html; charset=utf-8", []byte("<b>Goravel</b>"))
	})

	code, body, _, _ := s.request("GET", "/data", nil)

	s.Equal("<b>Goravel</b>", body)
	s.Equal(http.StatusOK, code)
}

func (s *ContextResponseSuite) TestSuccess_Json() {
	s.route.Get("/json", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Success().Json(contractshttp.Json{
			"id": "1",
		})
	})

	code, body, _, _ := s.request("GET", "/json", nil)

	s.Equal("{\"id\":\"1\"}", body)
	s.Equal(http.StatusOK, code)
}

func (s *ContextResponseSuite) TestSuccess_String() {
	s.route.Get("/string", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Success().String("Goravel")
	})

	code, body, _, _ := s.request("GET", "/string", nil)

	s.Equal("Goravel", body)
	s.Equal(http.StatusOK, code)
}

func (s *ContextResponseSuite) TestStatus_Data() {
	s.route.Get("/data", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Status(http.StatusCreated).Data("text/html; charset=utf-8", []byte("<b>Goravel</b>"))
	})

	code, body, _, _ := s.request("GET", "/data", nil)

	s.Equal("<b>Goravel</b>", body)
	s.Equal(http.StatusCreated, code)
}

func (s *ContextResponseSuite) TestStatus_Json() {
	s.route.Get("/json", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Status(http.StatusCreated).Json(contractshttp.Json{
			"id": "1",
		})
	})

	code, body, _, _ := s.request("GET", "/json", nil)

	s.Equal("{\"id\":\"1\"}", body)
	s.Equal(http.StatusCreated, code)
}

func (s *ContextResponseSuite) TestStatus_String() {
	s.route.Get("/string", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Status(http.StatusCreated).String("Goravel")
	})

	code, body, _, _ := s.request("GET", "/string", nil)

	s.Equal("Goravel", body)
	s.Equal(http.StatusCreated, code)
}

func (s *ContextResponseSuite) TestWithoutCookie() {
	cookieName := "goravel"
	s.route.Get("/without-cookie", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().WithoutCookie(cookieName).String(http.StatusOK, "Goravel")
	})

	code, body, _, cookies := s.request("GET", "/without-cookie", nil)

	s.Equal("Goravel", body)
	s.Equal(http.StatusOK, code)
	exist := false
	for _, cookie := range cookies {
		if cookie.Name == cookieName {
			exist = true
			s.Empty(cookie.Value)
		}
	}
	s.True(exist)
}

func (s *ContextResponseSuite) TestAbort() {
	s.route.Get("/abort-redirected", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().String(http.StatusOK, "redirected")
	})

	s.route.Middleware(testJson()).Get("/json/abort", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().String(http.StatusOK, "Goravel")
	})

	s.route.Middleware(testRedirect()).Get("/redirect/abort", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().String(http.StatusOK, "redirect")
	})

	code, body, _, _ := s.request("GET", "/json/abort", nil)
	s.Equal(contractshttp.StatusOK, code)
	s.JSONEq(`{"name":"abort json"}`, body)

	code, _, headers, _ := s.request("GET", "/redirect/abort", nil)
	s.Equal(contractshttp.StatusMovedPermanently, code)
	s.Equal("/abort-redirected", headers.Get("Location"))
}

func (s *ContextResponseSuite) request(method, url string, body io.Reader) (int, string, http.Header, []*http.Cookie) {
	req, err := http.NewRequest(method, url, body)
	s.Require().Nil(err)

	w := httptest.NewRecorder()
	s.route.ServeHTTP(w, req)

	return w.Code, w.Body.String(), w.Header(), w.Result().Cookies()
}

type testJsonMiddleware struct{}

func (m *testJsonMiddleware) Signature() string { return "testJson" }
func (m *testJsonMiddleware) Handle(ctx contractshttp.Context) {
	err := ctx.Response().Json(contractshttp.StatusOK, map[string]any{
		"name": "abort json",
	}).Abort()
	if err != nil {
		panic(err)
	}
	ctx.Request().Next()
}

type testRedirectMiddleware struct{}

func (m *testRedirectMiddleware) Signature() string { return "testRedirect" }
func (m *testRedirectMiddleware) Handle(ctx contractshttp.Context) {
	err := ctx.Response().Redirect(contractshttp.StatusMovedPermanently, "/abort-redirected").Abort()
	if err != nil {
		panic(err)
	}
	ctx.Request().Next()
}

type contextResponseTestMiddleware struct {
	s *ContextResponseSuite
}

func (m *contextResponseTestMiddleware) Signature() string {
	return "contextResponseTest"
}
func (m *contextResponseTestMiddleware) Handle(ctx contractshttp.Context) {
	ctx.Response().Header("global", "goravel")
	ctx.Request().Next()

	m.s.Equal("Goravel", ctx.Response().Origin().Body().String())
	m.s.Equal("goravel", ctx.Response().Origin().Header().Get("global"))
	m.s.Equal(7, ctx.Response().Origin().Size())
	m.s.Equal(http.StatusOK, ctx.Response().Origin().Status())
}

func testJson() contractshttp.Middleware {
	return &testJsonMiddleware{}
}

func testRedirect() contractshttp.Middleware {
	return &testRedirectMiddleware{}
}

func TestRemoveSetCookie(t *testing.T) {
	tests := []struct {
		name       string
		existing   []string
		cookieName string
		expect     []string
	}{
		{
			name:       "no headers",
			cookieName: "session",
		},
		{
			name:       "removes only the matching name",
			existing:   []string{"session=old; Path=/", "other=kept; Path=/"},
			cookieName: "session",
			expect:     []string{"other=kept; Path=/"},
		},
		{
			name:       "keeps names sharing a prefix",
			existing:   []string{"session=old", "session2=kept", "xsession=kept"},
			cookieName: "session",
			expect:     []string{"session2=kept", "xsession=kept"},
		},
		{
			name:       "removes every occurrence",
			existing:   []string{"session=a", "session=b"},
			cookieName: "session",
		},
		{
			name:       "keeps values that fail to parse",
			existing:   []string{"not a set-cookie", "session=old"},
			cookieName: "session",
			expect:     []string{"not a set-cookie"},
		},
		{
			name:       "no match leaves headers untouched",
			existing:   []string{"other=kept"},
			cookieName: "session",
			expect:     []string{"other=kept"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := http.Header{}
			for _, value := range tt.existing {
				header.Add("Set-Cookie", value)
			}

			removeSetCookie(header, tt.cookieName)

			assert.Equal(t, tt.expect, header.Values("Set-Cookie"))
		})
	}
}
