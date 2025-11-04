package gin

import (
	"crypto/tls"
	"errors"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin/render"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
	configmocks "github.com/goravel/framework/mocks/config"
	mockslog "github.com/goravel/framework/mocks/log"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type RouteTestSuite struct {
	suite.Suite
	mockConfig *configmocks.Config
	mockLog    *mockslog.Log
	route      *Route
}

func TestRouteTestSuite(t *testing.T) {
	suite.Run(t, new(RouteTestSuite))
}

func (s *RouteTestSuite) SetupTest() {
	s.mockConfig = configmocks.NewConfig(s.T())
	s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	s.mockConfig.EXPECT().GetInt("http.drivers.gin.body_limit", 4096).Return(4096).Once()
	s.mockConfig.EXPECT().Get("http.drivers.gin.template").Return(nil).Once()
	ConfigFacade = s.mockConfig

	s.mockLog = mockslog.NewLog(s.T())
	LogFacade = s.mockLog

	s.route = &Route{
		config: s.mockConfig,
		driver: "gin",
	}
	s.Require().Nil(s.route.init(nil))

	routes = make(map[string]map[string]contractshttp.Info)
}

func (s *RouteTestSuite) TestRecover() {
	s.Run("default", func() {
		s.mockLog.EXPECT().WithContext(mock.AnythingOfType("*gin.Context")).Return(s.mockLog).Once()
		s.mockLog.EXPECT().Request(mock.AnythingOfType("*gin.ContextRequest")).Return(s.mockLog).Once()
		s.mockLog.EXPECT().Error(1).Return().Once()

		s.route.Get("/recover", func(ctx contractshttp.Context) contractshttp.Response {
			panic(1)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/recover", nil)
		s.route.ServeHTTP(w, req)

		s.Empty(w.Body.String())
		s.Equal(http.StatusInternalServerError, w.Code)
	})

	s.Run("with custom callback", func() {
		s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
		s.mockConfig.EXPECT().GetInt("http.drivers.gin.body_limit", 4096).Return(4096).Once()
		s.mockConfig.EXPECT().Get("http.drivers.gin.template").Return(nil).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/recover", nil)

		called := false
		globalRecover := func(ctx contractshttp.Context, err any) {
			called = true
			ctx.Request().Abort(http.StatusServiceUnavailable)
		}

		s.route.Recover(globalRecover)

		s.route.Get("/recover", func(ctx contractshttp.Context) contractshttp.Response {
			panic(1)
		})

		s.route.ServeHTTP(w, req)

		s.Empty(w.Body.String())
		s.Equal(http.StatusServiceUnavailable, w.Code)
		s.True(called)
	})
}

func (s *RouteTestSuite) TestFallback() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/fallback", nil)

	s.route.Fallback(func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().String(404, "not found")
	})

	s.route.ServeHTTP(w, req)

	s.Equal("not found", w.Body.String())
	s.Equal(http.StatusNotFound, w.Code)
}

func (s *RouteTestSuite) TestGetRoutes() {
	s.route.Get("/b/{id}", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().String(200, "ok")
	})
	s.route.Post("/b/{id}", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().String(200, "ok")
	})
	s.route.Get("/a/{id}", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().String(200, "ok")
	})

	routes := s.route.GetRoutes()
	s.Len(routes, 3)
	s.Equal("GET|HEAD", routes[0].Method)
	s.Equal("/a/{id}", routes[0].Path)
	s.Equal("GET|HEAD", routes[1].Method)
	s.Equal("/b/{id}", routes[1].Path)
	s.Equal("POST", routes[2].Method)
	s.Equal("/b/{id}", routes[2].Path)
}

func (s *RouteTestSuite) TestGlobalMiddleware() {
	s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	s.mockConfig.EXPECT().GetInt("http.drivers.gin.body_limit", 4096).Return(4096).Once()
	s.mockConfig.EXPECT().Get("http.drivers.gin.template").Return(nil).Once()

	middleware := func(ctx contractshttp.Context) {}
	s.route.GlobalMiddleware(middleware)
	s.Len(s.route.instance.Handlers, 3)
}

func (s *RouteTestSuite) TestListen() {
	s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	s.mockConfig.EXPECT().GetInt("http.drivers.gin.header_limit", 4096).Return(4096).Once()

	s.route.Get("/", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Json(200, contractshttp.Json{
			"Hello": "Goravel",
		})
	})

	go func() {
		l, err := net.Listen("tcp", "127.0.0.1:3102")
		s.Require().Nil(err)
		err = s.route.Listen(l)
		s.Require().Nil(err)
	}()

	time.Sleep(1 * time.Second)

	resp, err := http.Get("http://127.0.0.1:3102")
	s.Require().Nil(err)
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	s.Nil(err)
	s.Equal("{\"Hello\":\"Goravel\"}", string(body))
}

func (s *RouteTestSuite) TestListenTLS() {
	s.route.Get("/", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Json(200, contractshttp.Json{
			"Hello": "Goravel",
		})
	})

	s.mockConfig.EXPECT().GetInt("http.drivers.gin.header_limit", 4096).Return(4096).Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.cert").Return("test_ca.crt").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.key").Return("test_ca.key").Once()
	s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()

	go func() {
		l, err := net.Listen("tcp", "127.0.0.1:3103")
		s.Require().Nil(err)
		s.Require().Nil(s.route.ListenTLS(l))
	}()

	time.Sleep(1 * time.Second)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get("https://127.0.0.1:3103")
	s.Require().Nil(err)
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	s.Nil(err)
	s.Equal("{\"Hello\":\"Goravel\"}", string(body))
}

func (s *RouteTestSuite) TestListenTLSWithCert() {
	s.route.Get("/", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Json(200, contractshttp.Json{
			"Hello": "Goravel",
		})
	})

	s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	s.mockConfig.EXPECT().GetInt("http.drivers.gin.header_limit", 4096).Return(4096).Once()

	go func() {
		l, err := net.Listen("tcp", "127.0.0.1:3104")
		s.Require().Nil(err)
		s.Require().Nil(s.route.ListenTLSWithCert(l, "test_ca.crt", "test_ca.key"))
	}()

	time.Sleep(1 * time.Second)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get("https://127.0.0.1:3104")
	s.Require().Nil(err)
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	s.Nil(err)
	s.Equal("{\"Hello\":\"Goravel\"}", string(body))
}

func (s *RouteTestSuite) TestInfo() {
	s.route.Get("/test", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Json(200, contractshttp.Json{
			"Hello": "Goravel",
		})
	}).Name("test")

	info := s.route.Info("test")
	s.Equal("GET|HEAD", info.Method)
	s.Equal("test", info.Name)
	s.Equal("/test", info.Path)
}

func (s *RouteTestSuite) TestRun() {
	s.Run("error when default port is empty", func() {
		s.SetupTest()

		s.mockConfig.EXPECT().GetString("http.host").Return("127.0.0.1").Once()
		s.mockConfig.EXPECT().GetString("http.port").Return("").Once()

		err := s.route.Run()

		s.Equal(errors.New("port can't be empty"), err)
	})

	s.Run("use default host", func() {
		s.SetupTest()

		s.route.Get("/", func(ctx contractshttp.Context) contractshttp.Response {
			return ctx.Response().Json(200, contractshttp.Json{
				"Hello": "Goravel",
			})
		})

		host := "127.0.0.1"
		port := "3031"

		s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
		s.mockConfig.EXPECT().GetString("http.host").Return(host).Once()
		s.mockConfig.EXPECT().GetString("http.port").Return(port).Once()
		s.mockConfig.EXPECT().GetInt("http.drivers.gin.header_limit", 4096).Return(4096).Once()

		var err error

		go func() {
			err = s.route.Run()
		}()

		defer s.NoError(s.route.Shutdown())

		time.Sleep(1 * time.Second)

		s.NoError(err)

		hostUrl := "http://" + host + ":" + port
		resp, err := http.Get(hostUrl)
		s.Require().Nil(err)
		defer func() {
			_ = resp.Body.Close()
		}()

		body, err := io.ReadAll(resp.Body)
		s.Require().Nil(err)
		s.Equal("{\"Hello\":\"Goravel\"}", string(body))
	})

	s.Run("use custom host", func() {
		s.SetupTest()

		s.route.Get("/", func(ctx contractshttp.Context) contractshttp.Response {
			return ctx.Response().Json(200, contractshttp.Json{
				"Hello": "Goravel",
			})
		})

		s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
		s.mockConfig.EXPECT().GetInt("http.drivers.gin.header_limit", 4096).Return(4096).Once()

		var err error

		go func() {
			err = s.route.Run("127.0.0.1:3032")
		}()

		defer s.NoError(s.route.Shutdown())

		time.Sleep(1 * time.Second)

		s.NoError(err)

		hostUrl := "http://127.0.0.1:3032"
		resp, err := http.Get(hostUrl)
		s.Require().Nil(err)
		defer func() {
			_ = resp.Body.Close()
		}()

		body, err := io.ReadAll(resp.Body)
		s.Require().Nil(err)
		s.Equal("{\"Hello\":\"Goravel\"}", string(body))
	})
}

func (s *RouteTestSuite) TestRunTLS() {
	s.Run("error when default port is empty", func() {
		s.SetupTest()

		s.mockConfig.EXPECT().GetString("http.tls.host").Return("127.0.0.1").Once()
		s.mockConfig.EXPECT().GetString("http.tls.port").Return("").Once()

		err := s.route.RunTLS()

		s.Equal(errors.New("port can't be empty"), err)
	})

	s.Run("use default host", func() {
		s.SetupTest()

		s.route.Get("/", func(ctx contractshttp.Context) contractshttp.Response {
			return ctx.Response().Json(200, contractshttp.Json{
				"Hello": "Goravel",
			})
		})

		host := "127.0.0.1"
		port := "3033"
		addr := "https://" + host + ":" + port

		s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
		s.mockConfig.EXPECT().GetInt("http.drivers.gin.header_limit", 4096).Return(4096).Once()
		s.mockConfig.EXPECT().GetString("http.tls.host").Return(host).Once()
		s.mockConfig.EXPECT().GetString("http.tls.port").Return(port).Once()
		s.mockConfig.EXPECT().GetString("http.tls.ssl.cert").Return("test_ca.crt").Once()
		s.mockConfig.EXPECT().GetString("http.tls.ssl.key").Return("test_ca.key").Once()

		var err error

		go func() {
			err = s.route.RunTLS()
		}()

		defer s.NoError(s.route.Shutdown())

		time.Sleep(1 * time.Second)

		s.NoError(err)

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		resp, err := client.Get(addr)
		s.Require().Nil(err)
		defer func() {
			_ = resp.Body.Close()
		}()

		body, err := io.ReadAll(resp.Body)
		s.Require().Nil(err)
		s.Equal("{\"Hello\":\"Goravel\"}", string(body))
	})

	s.Run("use custom host", func() {
		s.SetupTest()

		s.route.Get("/", func(ctx contractshttp.Context) contractshttp.Response {
			return ctx.Response().Json(200, contractshttp.Json{
				"Hello": "Goravel",
			})
		})

		host := "127.0.0.1"
		port := "3034"
		addr := "https://" + host + ":" + port

		s.mockConfig.EXPECT().GetInt("http.drivers.gin.header_limit", 4096).Return(4096).Once()
		s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
		s.mockConfig.EXPECT().GetString("http.tls.ssl.cert").Return("test_ca.crt").Once()
		s.mockConfig.EXPECT().GetString("http.tls.ssl.key").Return("test_ca.key").Once()

		var err error

		go func() {
			err = s.route.RunTLS(host + ":" + port)
		}()

		defer s.NoError(s.route.Shutdown())

		time.Sleep(1 * time.Second)

		s.NoError(err)

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		resp, err := client.Get(addr)
		s.Require().Nil(err)
		defer func() {
			_ = resp.Body.Close()
		}()

		body, err := io.ReadAll(resp.Body)
		s.Require().Nil(err)
		s.Equal("{\"Hello\":\"Goravel\"}", string(body))
	})
}

func (s *RouteTestSuite) TestRunTLSWithCert() {
	s.Run("error when default host is empty", func() {
		s.SetupTest()

		err := s.route.RunTLSWithCert("", "test_ca.crt", "test_ca.key")

		s.Equal(errors.New("host can't be empty"), err)
	})

	s.Run("error when certificate is empty", func() {
		s.SetupTest()

		err := s.route.RunTLSWithCert("127.0.0.1:3032", "", "test_ca.key")

		time.Sleep(1 * time.Second)

		s.Equal(errors.New("certificate can't be empty"), err)
	})

	s.Run("error when key is empty", func() {
		s.SetupTest()

		err := s.route.RunTLSWithCert("127.0.0.1:3032", "test_ca.crt", "")

		time.Sleep(1 * time.Second)

		s.Equal(errors.New("certificate can't be empty"), err)
	})

	s.Run("happy path", func() {
		s.SetupTest()

		s.route.Get("/", func(ctx contractshttp.Context) contractshttp.Response {
			return ctx.Response().Json(200, contractshttp.Json{
				"Hello": "Goravel",
			})
		})

		s.mockConfig.EXPECT().GetInt("http.drivers.gin.header_limit", 4096).Return(4096).Once()
		s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()

		var (
			host = "127.0.0.1"
			port = "3035"
			addr = "https://" + host + ":" + port

			err error
		)

		go func() {
			err = s.route.RunTLSWithCert(host+":"+port, "test_ca.crt", "test_ca.key")
		}()

		defer s.NoError(s.route.Shutdown())

		time.Sleep(1 * time.Second)

		s.NoError(err)

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		resp, err := client.Get(addr)
		s.Require().Nil(err)
		defer func() {
			_ = resp.Body.Close()
		}()

		body, err := io.ReadAll(resp.Body)
		s.Require().Nil(err)
		s.Equal("{\"Hello\":\"Goravel\"}", string(body))
	})
}

func (s *RouteTestSuite) TestNewRoute() {
	defaultTemplate, err := DefaultTemplate()
	s.Require().Nil(err)

	tests := []struct {
		name        string
		parameters  map[string]any
		setup       func()
		expectError error
	}{
		{
			name:        "parameters is nil",
			setup:       func() {},
			expectError: errors.New("please set the driver"),
		},
		{
			name:       "template is instance",
			parameters: map[string]any{"driver": "gin"},
			setup: func() {
				s.mockConfig.EXPECT().GetInt("http.request_timeout", 3).Return(3).Once()
				s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
				s.mockConfig.EXPECT().GetInt("http.drivers.gin.body_limit", 4096).Return(4096).Once()
				s.mockConfig.EXPECT().Get("http.drivers.gin.template").Return(defaultTemplate).Once()
			},
		},
		{
			name:       "template is callback and returns success",
			parameters: map[string]any{"driver": "gin"},
			setup: func() {
				s.mockConfig.EXPECT().GetInt("http.request_timeout", 3).Return(3).Once()
				s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
				s.mockConfig.EXPECT().GetInt("http.drivers.gin.body_limit", 4096).Return(4096).Once()
				s.mockConfig.EXPECT().Get("http.drivers.gin.template").Return(func() (render.HTMLRender, error) {
					return defaultTemplate, nil
				}).Once()
			},
		},
		{
			name:       "template is callback and returns error",
			parameters: map[string]any{"driver": "gin"},
			setup: func() {
				s.mockConfig.EXPECT().GetInt("http.request_timeout", 3).Return(3).Once()
				s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
				s.mockConfig.EXPECT().GetInt("http.drivers.gin.body_limit", 4096).Return(4096).Once()
				s.mockConfig.EXPECT().Get("http.drivers.gin.template").Return(func() (render.HTMLRender, error) {
					return nil, errors.New("error")
				}).Once()
			},
			expectError: errors.New("error"),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.setup()

			route, err := NewRoute(s.mockConfig, tt.parameters)

			if tt.expectError != nil {
				s.Equal(tt.expectError, err)
			} else {
				s.NoError(err)
				s.NotNil(route)
			}
		})
	}
}

func (s *RouteTestSuite) TestShutdown() {
	host := "127.0.0.1"
	port := "3036"
	addr := "http://" + host + ":" + port

	s.Run("no new requests will be accepted after shutdown", func() {
		s.SetupTest()

		s.route.Get("/", func(ctx contractshttp.Context) contractshttp.Response {
			return ctx.Response().Success().String("Goravel")
		})

		s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
		s.mockConfig.EXPECT().GetInt("http.drivers.gin.header_limit", 4096).Return(4096).Once()
		s.mockConfig.EXPECT().GetString("http.host").Return(host).Once()
		s.mockConfig.EXPECT().GetString("http.port").Return(port).Once()

		var err error

		go func() {
			err = s.route.Run()
		}()

		defer s.NoError(s.route.Shutdown())

		time.Sleep(1 * time.Second)

		s.NoError(err)

		assertHttpNormal(s.T(), addr, true)

		s.NoError(s.route.Shutdown())

		assertHttpNormal(s.T(), addr, false)
	})

	s.Run("ensure that received requests are processed", func() {
		var (
			count atomic.Int64
			err   error
		)

		s.SetupTest()

		s.route.Get("/", func(ctx contractshttp.Context) contractshttp.Response {
			time.Sleep(time.Second)
			defer count.Add(1)
			return ctx.Response().Success().String("Goravel")
		})

		s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
		s.mockConfig.EXPECT().GetInt("http.drivers.gin.header_limit", 4096).Return(4096).Once()
		s.mockConfig.EXPECT().GetString("http.host").Return(host).Once()
		s.mockConfig.EXPECT().GetString("http.port").Return(port).Once()

		go func() {
			err = s.route.Run()
		}()

		time.Sleep(1 * time.Second)

		s.NoError(err)

		wg := sync.WaitGroup{}
		count.Store(0)
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				assertHttpNormal(s.T(), addr, true)
			}()
		}

		time.Sleep(100 * time.Millisecond)

		s.NoError(s.route.Shutdown())

		assertHttpNormal(s.T(), addr, false)

		wg.Wait()
		s.Equal(int64(3), count.Load())
	})
}

func (s *RouteTestSuite) TestTest() {
	s.route.Get("/", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Success().String("Hello, Goravel!")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	resp, err := s.route.Test(req)

	s.NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	s.NoError(err)
	s.Equal("Hello, Goravel!", string(body))
}

func assertHttpNormal(t *testing.T, addr string, expectNormal bool) {
	resp, err := http.DefaultClient.Get(addr)
	if !expectNormal {
		assert.NotNil(t, err)
		assert.Nil(t, resp)
	} else {
		assert.Nil(t, err)
		assert.NotNil(t, resp)
		if resp != nil {
			assert.Equal(t, resp.StatusCode, http.StatusOK)
			body, err := io.ReadAll(resp.Body)
			assert.Nil(t, err)
			assert.Equal(t, string(body), "Goravel")
		}
	}
}

type CreateUser struct {
	Name string `form:"name" json:"name"`
}

func (r *CreateUser) Authorize(ctx contractshttp.Context) error {
	return nil
}

func (r *CreateUser) Rules(ctx contractshttp.Context) map[string]string {
	return map[string]string{
		"name": "required",
	}
}

func (r *CreateUser) Filters(ctx contractshttp.Context) map[string]string {
	return map[string]string{
		"name": "trim",
	}
}

func (r *CreateUser) Messages(ctx contractshttp.Context) map[string]string {
	return map[string]string{}
}

func (r *CreateUser) Attributes(ctx contractshttp.Context) map[string]string {
	return map[string]string{}
}

func (r *CreateUser) PrepareForValidation(ctx contractshttp.Context, data validation.Data) error {
	test := cast.ToString(ctx.Value("test"))
	if name, exist := data.Get("name"); exist {
		return data.Set("name", name.(string)+"1"+test)
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

func (r *Unauthorize) Filters(ctx contractshttp.Context) map[string]string {
	return map[string]string{
		"name": "trim",
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

type FileImageJson struct {
	Name  string                 `form:"name" json:"name"`
	File  *multipart.FileHeader  `form:"file" json:"file"`
	Files []multipart.FileHeader `form:"files" json:"files"`
	Image *multipart.FileHeader  `form:"image" json:"image"`
	Json  string                 `form:"json" json:"json"`
}

func (r *FileImageJson) Authorize(ctx contractshttp.Context) error {
	return nil
}

func (r *FileImageJson) Rules(ctx contractshttp.Context) map[string]string {
	return map[string]string{
		"name":  "required",
		"file":  "file",
		"image": "image",
		"json":  "json",
	}
}

func (r *FileImageJson) Filters(ctx contractshttp.Context) map[string]string {
	return map[string]string{
		"name": "trim",
	}
}

func (r *FileImageJson) Messages(ctx contractshttp.Context) map[string]string {
	return map[string]string{}
}

func (r *FileImageJson) Attributes(ctx contractshttp.Context) map[string]string {
	return map[string]string{}
}

func (r *FileImageJson) PrepareForValidation(ctx contractshttp.Context, data validation.Data) error {
	return nil
}
