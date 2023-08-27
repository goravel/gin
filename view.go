package gin

import (
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
)

type View struct {
	instance *gin.Context
}

func NewView(instance *gin.Context) *View {
	return &View{instance: instance}
}

func (receive *View) Make(view string, data ...any) {
	shared := ViewFacade.GetShared()
	if len(data) == 0 {
		receive.instance.HTML(200, view, shared)
	} else {
		dataMap := make(map[string]any)
		dataType := reflect.TypeOf(data[0])
		switch dataType.Kind() {
		case reflect.Struct:
			dataMap = structToMap(data[0])
			for key, value := range dataMap {
				shared[key] = value
			}
			receive.instance.HTML(200, view, shared)
		case reflect.Map:
			item := data[0]
			dataValue := reflect.ValueOf(item)
			keys := dataValue.MapKeys()
			for key, value := range shared {
				exist := false
				for _, k := range keys {
					if k.String() == key {
						exist = true
						break
					}
				}
				if !exist {
					dataValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
				}
			}
			receive.instance.HTML(200, view, item)
		default:
			panic(fmt.Sprintf("make %s view failed, data must be map[string]any or struct", view))
		}
	}
}

func (receive *View) First(views []string, data ...any) {
	for _, view := range views {
		if ViewFacade.Exists(view) {
			receive.Make(view, data...)
			return
		}
	}
}

func structToMap(data any) map[string]any {
	res := make(map[string]any)
	modelType := reflect.TypeOf(data)
	modelValue := reflect.ValueOf(data)

	if modelType.Kind() == reflect.Pointer {
		modelType = modelType.Elem()
		modelValue = modelValue.Elem()
	}

	for i := 0; i < modelType.NumField(); i++ {
		dbColumn := modelType.Field(i).Name
		if modelValue.Field(i).Kind() == reflect.Pointer {
			if modelValue.Field(i).IsNil() {
				res[dbColumn] = nil
			} else {
				res[dbColumn] = modelValue.Field(i).Elem().Interface()
			}
		} else {
			res[dbColumn] = modelValue.Field(i).Interface()
		}
	}

	return res
}
