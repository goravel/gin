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

func (r *DataResponse) Render() error {
	r.instance.Data(r.code, r.contentType, r.data)

	return nil
}

type DownloadResponse struct {
	filename string
	filepath string
	instance *gin.Context
}

func (r *DownloadResponse) Render() error {
	r.instance.FileAttachment(r.filepath, r.filename)

	return nil
}

type FileResponse struct {
	filepath string
	instance *gin.Context
}

func (r *FileResponse) Render() error {
	r.instance.File(r.filepath)

	return nil
}

type JsonResponse struct {
	code     int
	obj      any
	instance *gin.Context
}

func (r *JsonResponse) Render() error {
	r.instance.JSON(r.code, r.obj)

	return nil
}

type RedirectResponse struct {
	code     int
	location string
	instance *gin.Context
}

func (r *RedirectResponse) Render() error {
	r.instance.Redirect(r.code, r.location)

	return nil
}

type StringResponse struct {
	code     int
	format   string
	instance *gin.Context
	values   []any
}

func (r *StringResponse) Render() error {
	r.instance.String(r.code, r.format, r.values...)

	return nil
}

type HtmlResponse struct {
	data     any
	instance *gin.Context
	view     string
}

func (r *HtmlResponse) Render() error {
	r.instance.HTML(http.StatusOK, r.view, r.data)

	return nil
}
