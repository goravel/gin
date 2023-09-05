package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type DataResponse struct {
	code        int
	contentType string
	data        []byte
	instance    *gin.Context
}

func (r *DataResponse) Render() {
	r.instance.Data(r.code, r.contentType, r.data)
}

type DownloadResponse struct {
	filename string
	filepath string
	instance *gin.Context
}

func (r *DownloadResponse) Render() {
	r.instance.FileAttachment(r.filepath, r.filename)
}

type FileResponse struct {
	filepath string
	instance *gin.Context
}

func (r *FileResponse) Render() {
	r.instance.File(r.filepath)
}

type JsonResponse struct {
	code     int
	obj      any
	instance *gin.Context
}

func (r *JsonResponse) Render() {
	r.instance.JSON(r.code, r.obj)
}

type RedirectResponse struct {
	code     int
	location string
	instance *gin.Context
}

func (r *RedirectResponse) Render() {
	r.instance.Redirect(r.code, r.location)
}

type StringResponse struct {
	code     int
	format   string
	instance *gin.Context
	values   []any
}

func (r *StringResponse) Render() {
	r.instance.String(r.code, r.format, r.values...)
}

type HtmlResponse struct {
	data     any
	instance *gin.Context
	view     string
}

func (r *HtmlResponse) Render() {
	r.instance.HTML(http.StatusOK, r.view, r.data)
}
