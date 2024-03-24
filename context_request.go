package gin

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gookit/validate"
	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
	contractsvalidate "github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/filesystem"
	"github.com/goravel/framework/support/json"
	"github.com/goravel/framework/validation"
	"github.com/spf13/cast"
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

	for _, param := range r.instance.Params {
		dataMap[param.Key] = param.Value
	}
	for k, v := range queryMap {
		dataMap[k] = v
	}
	for k, v := range r.postData {
		dataMap[k] = v
	}

	return dataMap
}

func (r *ContextRequest) Bind(obj any) error {
	return r.instance.ShouldBind(obj)
}

func (r *ContextRequest) Cookie(key string, defaultValue ...string) string {
	cookie, err := r.instance.Cookie(key)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}

		return ""
	}

	return cookie
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
	valueFromPostData := r.getValueFromPostData(key)
	if valueFromPostData != nil {
		switch reflect.ValueOf(valueFromPostData).Kind() {
		case reflect.Map:
			valueFromPostDataObByte, err := json.Marshal(valueFromPostData)
			if err != nil {
				return ""
			}

			return string(valueFromPostDataObByte)
		case reflect.Slice:
			return strings.Join(cast.ToStringSlice(valueFromPostData), ",")
		default:
			return cast.ToString(valueFromPostData)
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
	if valueFromPostData := r.getValueFromPostData(key); valueFromPostData != nil {
		return cast.ToStringSlice(valueFromPostData)
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	} else {
		return []string{}
	}
}

func (r *ContextRequest) InputMap(key string, defaultValue ...map[string]string) map[string]string {
	if valueFromPostData := r.getValueFromPostData(key); valueFromPostData != nil {
		return cast.ToStringMapString(valueFromPostData)
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
	dataFace := validate.FromMap(r.ctx.Request().All())

	if generateOptions["prepareForValidation"] != nil {
		if err := generateOptions["prepareForValidation"].(func(ctx contractshttp.Context, data contractsvalidate.Data) error)(r.ctx, validation.NewData(dataFace)); err != nil {
			return nil, err
		}
	}

	v = dataFace.Create()

	if generateOptions["filters"] != nil {
		v.FilterRules(generateOptions["filters"].(map[string]string))
	}

	validation.AppendOptions(v, generateOptions)

	return validation.NewValidator(v), nil
}

func (r *ContextRequest) ValidateRequest(request contractshttp.FormRequest) (contractsvalidate.Errors, error) {
	if err := request.Authorize(r.ctx); err != nil {
		return nil, err
	}

	filters := make(map[string]string)
	val := reflect.Indirect(reflect.ValueOf(request))
	for i := 0; i < val.Type().NumField(); i++ {
		field := val.Type().Field(i)
		form := field.Tag.Get("form")
		filter := field.Tag.Get("filter")
		if len(form) > 0 && len(filter) > 0 {
			filters[form] = filter
		}
	}

	validator, err := r.Validate(request.Rules(r.ctx), validation.Messages(request.Messages(r.ctx)), validation.Attributes(request.Attributes(r.ctx)), func(options map[string]any) {
		options["prepareForValidation"] = request.PrepareForValidation
		options["filters"] = filters
	})
	if err != nil {
		return nil, err
	}

	if err := validator.Bind(request); err != nil {
		return nil, err
	}

	return validator.Errors(), nil
}

func (r *ContextRequest) getValueFromPostData(key string) any {
	if r.postData == nil {
		return nil
	}

	var current any
	current = r.postData
	keys := strings.Split(key, ".")
	for _, k := range keys {
		currentValue := reflect.ValueOf(current)
		switch currentValue.Kind() {
		case reflect.Map:
			if value := currentValue.MapIndex(reflect.ValueOf(k)); value.IsValid() {
				current = value.Interface()
			} else {
				if value := currentValue.MapIndex(reflect.ValueOf(k + "[]")); value.IsValid() {
					current = value.Interface()
				} else {
					return nil
				}
			}
		case reflect.Slice:
			if number, err := strconv.Atoi(k); err == nil {
				return cast.ToStringSlice(current)[number]
			} else {
				return nil
			}
		}
	}

	return current
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

		if err := json.Unmarshal(bodyBytes, &data); err != nil {
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
			if len(v) > 1 {
				data[k] = v
			} else if len(v) == 1 {
				data[k] = v[0]
			}
		}
		for k, v := range request.MultipartForm.File {
			if len(v) > 1 {
				data[k] = v
			} else if len(v) == 1 {
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
			if len(v) > 1 {
				data[k] = v
			} else if len(v) == 1 {
				data[k] = v[0]
			}
		}
	}

	return data, nil
}

func stringToBool(value string) bool {
	return value == "1" || value == "true" || value == "on" || value == "yes"
}
