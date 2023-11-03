package gin

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/gookit/validate"
	"github.com/spf13/cast"

	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
	contractsvalidate "github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/filesystem"
	"github.com/goravel/framework/validation"
)

type ContextRequest struct {
	ctx        *Context
	instance   *gin.Context
	postData   map[string]any
	log        log.Log
	validation contractsvalidate.Validation
}

func NewContextRequest(ctx *Context, log log.Log, validation contractsvalidate.Validation) contractshttp.ContextRequest {
	postData, err := getPostData(ctx)
	if err != nil {
		LogFacade.Error(fmt.Sprintf("%+v", errors.Unwrap(err)))
	}

	return &ContextRequest{ctx: ctx, instance: ctx.instance, postData: postData, log: log, validation: validation}
}

func (r *ContextRequest) AbortWithStatus(code int) {
	r.instance.AbortWithStatus(code)
}

func (r *ContextRequest) AbortWithStatusJson(code int, jsonObj any) {
	r.instance.AbortWithStatusJSON(code, jsonObj)
}

func (r *ContextRequest) All() map[string]any {
	var (
		dataMap  = make(map[string]any)
		queryMap = make(map[string]any)
	)

	for key, query := range r.instance.Request.URL.Query() {
		queryMap[key] = strings.Join(query, ",")
	}

	var mu sync.RWMutex
	for k, v := range queryMap {
		mu.Lock()
		dataMap[k] = v
		mu.Unlock()
	}
	for k, v := range r.postData {
		mu.Lock()
		dataMap[k] = v
		mu.Unlock()
	}

	return dataMap
}

func (r *ContextRequest) Bind(obj any) error {
	return r.instance.ShouldBind(obj)
}

func (r *ContextRequest) Form(key string, defaultValue ...string) string {
	if len(defaultValue) == 0 {
		return r.instance.PostForm(key)
	}

	return r.instance.DefaultPostForm(key, defaultValue[0])
}

func (r *ContextRequest) File(name string) (contractsfilesystem.File, error) {
	file, err := r.instance.FormFile(name)
	if err != nil {
		return nil, err
	}

	return filesystem.NewFileFromRequest(file)
}

func (r *ContextRequest) FullUrl() string {
	prefix := "https://"
	if r.instance.Request.TLS == nil {
		prefix = "http://"
	}

	if r.instance.Request.Host == "" {
		return ""
	}

	return prefix + r.instance.Request.Host + r.instance.Request.RequestURI
}

func (r *ContextRequest) Header(key string, defaultValue ...string) string {
	header := r.instance.GetHeader(key)
	if header != "" {
		return header
	}

	if len(defaultValue) == 0 {
		return ""
	}

	return defaultValue[0]
}

func (r *ContextRequest) Headers() http.Header {
	return r.instance.Request.Header
}

func (r *ContextRequest) Host() string {
	return r.instance.Request.Host
}

func (r *ContextRequest) Json(key string, defaultValue ...string) string {
	var data map[string]any
	if err := r.Bind(&data); err != nil {
		if len(defaultValue) == 0 {
			return ""
		} else {
			return defaultValue[0]
		}
	}

	if value, exist := data[key]; exist {
		return cast.ToString(value)
	}

	if len(defaultValue) == 0 {
		return ""
	}

	return defaultValue[0]
}

func (r *ContextRequest) Method() string {
	return r.instance.Request.Method
}

func (r *ContextRequest) Next() {
	r.instance.Next()
}

func (r *ContextRequest) Query(key string, defaultValue ...string) string {
	if len(defaultValue) > 0 {
		return r.instance.DefaultQuery(key, defaultValue[0])
	}

	return r.instance.Query(key)
}

func (r *ContextRequest) QueryInt(key string, defaultValue ...int) int {
	if val, ok := r.instance.GetQuery(key); ok {
		return cast.ToInt(val)
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return 0
}

func (r *ContextRequest) QueryInt64(key string, defaultValue ...int64) int64 {
	if val, ok := r.instance.GetQuery(key); ok {
		return cast.ToInt64(val)
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return 0
}

func (r *ContextRequest) QueryBool(key string, defaultValue ...bool) bool {
	if value, ok := r.instance.GetQuery(key); ok {
		return stringToBool(value)
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return false
}

func (r *ContextRequest) QueryArray(key string) []string {
	return r.instance.QueryArray(key)
}

func (r *ContextRequest) QueryMap(key string) map[string]string {
	return r.instance.QueryMap(key)
}

func (r *ContextRequest) Queries() map[string]string {
	queries := make(map[string]string)

	for key, query := range r.instance.Request.URL.Query() {
		queries[key] = strings.Join(query, ",")
	}

	return queries
}

func (r *ContextRequest) Origin() *http.Request {
	return r.instance.Request
}

func (r *ContextRequest) Path() string {
	return r.instance.Request.URL.Path
}

func (r *ContextRequest) Input(key string, defaultValue ...string) string {
	keys := strings.Split(key, ".")
	current := r.postData
	for _, k := range keys {
		value, found := current[k]
		if found {
			if nestedMap, isMap := value.(map[string]any); isMap {
				current = nestedMap
			} else {
				return cast.ToString(value)
			}
		}
	}

	if r.instance.Query(key) != "" {
		return r.instance.Query(key)
	}

	value := r.instance.Param(key)
	if len(value) == 0 && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return value
}

func (r *ContextRequest) InputArray(key string, defaultValue ...[]string) []string {
	keys := strings.Split(key, ".")
	current := r.postData
	for _, k := range keys {
		value, found := current[k]
		if !found {
			return []string{}
		}
		if nestedMap, isMap := value.(map[string]any); isMap {
			current = nestedMap
		} else {
			return cast.ToStringSlice(value)
		}
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	} else {
		return []string{}
	}
}

func (r *ContextRequest) InputMap(key string, defaultValue ...map[string]string) map[string]string {
	keys := strings.Split(key, ".")
	current := r.postData
	for _, k := range keys {
		value, found := current[k]
		if !found {
			return map[string]string{}
		}
		if nestedMap, isMap := value.(map[string]string); isMap {
			current = cast.ToStringMap(nestedMap)
		} else {
			return cast.ToStringMapString(value)
		}
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	} else {
		return map[string]string{}
	}
}

func (r *ContextRequest) InputInt(key string, defaultValue ...int) int {
	value := r.Input(key)
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return cast.ToInt(value)
}

func (r *ContextRequest) InputInt64(key string, defaultValue ...int64) int64 {
	value := r.Input(key)
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return cast.ToInt64(value)
}

func (r *ContextRequest) InputBool(key string, defaultValue ...bool) bool {
	value := r.Input(key)
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return stringToBool(value)
}

func (r *ContextRequest) Ip() string {
	return r.instance.ClientIP()
}

func (r *ContextRequest) Route(key string) string {
	return r.instance.Param(key)
}

func (r *ContextRequest) RouteInt(key string) int {
	val := r.instance.Param(key)

	return cast.ToInt(val)
}

func (r *ContextRequest) RouteInt64(key string) int64 {
	val := r.instance.Param(key)

	return cast.ToInt64(val)
}

func (r *ContextRequest) Url() string {
	return r.instance.Request.RequestURI
}

func (r *ContextRequest) Validate(rules map[string]string, options ...contractsvalidate.Option) (contractsvalidate.Validator, error) {
	if len(rules) == 0 {
		return nil, errors.New("rules can't be empty")
	}

	options = append(options, validation.Rules(rules), validation.CustomRules(r.validation.Rules()))
	generateOptions := validation.GenerateOptions(options)

	var v *validate.Validation
	dataFace, err := validate.FromRequest(r.Origin())
	if err != nil {
		return nil, err
	}

	for _, param := range r.instance.Params {
		if _, exist := dataFace.Get(param.Key); !exist {
			if _, err := dataFace.Set(param.Key, param.Value); err != nil {
				return nil, err
			}
		}
	}

	if generateOptions["prepareForValidation"] != nil {
		if err := generateOptions["prepareForValidation"].(func(ctx contractshttp.Context, data contractsvalidate.Data) error)(r.ctx, validation.NewData(dataFace)); err != nil {
			return nil, err
		}
	}

	v = dataFace.Create()
	validation.AppendOptions(v, generateOptions)

	return validation.NewValidator(v, dataFace), nil
}

func (r *ContextRequest) ValidateRequest(request contractshttp.FormRequest) (contractsvalidate.Errors, error) {
	if err := request.Authorize(r.ctx); err != nil {
		return nil, err
	}

	validator, err := r.Validate(request.Rules(r.ctx), validation.Messages(request.Messages(r.ctx)), validation.Attributes(request.Attributes(r.ctx)), func(options map[string]any) {
		options["prepareForValidation"] = request.PrepareForValidation
	})
	if err != nil {
		return nil, err
	}

	if err := validator.Bind(request); err != nil {
		return nil, err
	}

	return validator.Errors(), nil
}

func getPostData(ctx *Context) (map[string]any, error) {
	request := ctx.instance.Request
	if request == nil || request.Body == nil || request.ContentLength == 0 {
		return nil, nil
	}

	contentType := ctx.instance.ContentType()
	data := make(map[string]any)
	if contentType == "application/json" {
		bodyBytes, err := io.ReadAll(request.Body)
		_ = request.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("retrieve json error: %v", err)
		}

		if err := sonic.Unmarshal(bodyBytes, &data); err != nil {
			return nil, fmt.Errorf("decode json [%v] error: %v", string(bodyBytes), err)
		}

		request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	if contentType == "multipart/form-data" {
		if request.PostForm == nil {
			const defaultMemory = 32 << 20
			if err := request.ParseMultipartForm(defaultMemory); err != nil {
				return nil, fmt.Errorf("parse multipart form error: %v", err)
			}
		}
		for k, v := range request.PostForm {
			data[k] = strings.Join(v, ",")
		}
		for k, v := range request.MultipartForm.File {
			if len(v) > 0 {
				data[k] = v[0]
			}
		}
	}

	if contentType == "application/x-www-form-urlencoded" {
		if request.PostForm == nil {
			if err := request.ParseForm(); err != nil {
				return nil, fmt.Errorf("parse form error: %v", err)
			}
		}
		for k, v := range request.PostForm {
			data[k] = strings.Join(v, ",")
		}
	}

	return data, nil
}

func stringToBool(value string) bool {
	return value == "1" || value == "true" || value == "on" || value == "yes"
}
