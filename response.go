package gin

import (
	"bytes"
	"net/http"

	"github.com/gin-gonic/gin"

	httpcontract "github.com/goravel/framework/contracts/http"
)

type Response struct {
	instance *gin.Context
	origin   httpcontract.ResponseOrigin
}

func NewResponse(instance *gin.Context, origin httpcontract.ResponseOrigin) *Response {
	return &Response{instance, origin}
}

func (r *Response) Data(code int, contentType string, data []byte) {
	r.instance.Data(code, contentType, data)
}

func (r *Response) Download(filepath, filename string) {
	r.instance.FileAttachment(filepath, filename)
}

func (r *Response) File(filepath string) {
	r.instance.File(filepath)
}

func (r *Response) Header(key, value string) httpcontract.Response {
	r.instance.Header(key, value)

	return r
}

func (r *Response) Json(code int, obj any) {
	r.instance.JSON(code, obj)
}

func (r *Response) Origin() httpcontract.ResponseOrigin {
	return r.origin
}

func (r *Response) Redirect(code int, location string) {
	r.instance.Redirect(code, location)
}

func (r *Response) String(code int, format string, values ...any) {
	r.instance.String(code, format, values...)
}

func (r *Response) Success() httpcontract.ResponseSuccess {
	return NewGinSuccess(r.instance)
}

func (r *Response) Status(code int) httpcontract.ResponseStatus {
	return NewStatus(r.instance, code)
}

func (r *Response) View() httpcontract.ResponseView {
	return NewView(r.instance)
}

func (r *Response) Writer() http.ResponseWriter {
	return r.instance.Writer
}

func (r *Response) Flush() {
	r.instance.Writer.Flush()
}

type Success struct {
	instance *gin.Context
}

func NewGinSuccess(instance *gin.Context) httpcontract.ResponseSuccess {
	return &Success{instance}
}

func (r *Success) Data(contentType string, data []byte) {
	r.instance.Data(http.StatusOK, contentType, data)
}

func (r *Success) Json(obj any) {
	r.instance.JSON(http.StatusOK, obj)
}

func (r *Success) String(format string, values ...any) {
	r.instance.String(http.StatusOK, format, values...)
}

type Status struct {
	instance *gin.Context
	status   int
}

func NewStatus(instance *gin.Context, code int) httpcontract.ResponseSuccess {
	return &Status{instance, code}
}

func (r *Status) Data(contentType string, data []byte) {
	r.instance.Data(r.status, contentType, data)
}

func (r *Status) Json(obj any) {
	r.instance.JSON(r.status, obj)
}

func (r *Status) String(format string, values ...any) {
	r.instance.String(r.status, format, values...)
}

func ResponseMiddleware() httpcontract.Middleware {
	return func(ctx httpcontract.Context) {
		blw := &BodyWriter{body: bytes.NewBufferString("")}
		switch ctx := ctx.(type) {
		case *Context:
			blw.ResponseWriter = ctx.Instance().Writer
			ctx.Instance().Writer = blw
		}

		ctx.WithValue("responseOrigin", blw)
		ctx.Request().Next()
	}
}

type BodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *BodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)

	return w.ResponseWriter.Write(b)
}

func (w *BodyWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)

	return w.ResponseWriter.WriteString(s)
}

func (w *BodyWriter) Body() *bytes.Buffer {
	return w.body
}

func (w *BodyWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}
