package spi

import (
	"net/http"
	"reflect"
)

type IPMAASContainer interface {
	AddRoute(path string, handlerFunc http.HandlerFunc)
	RenderList(w http.ResponseWriter, r *http.Request, items []interface{})
	GetTemplate(templateInfo *TemplateInfo) (ITemplate, error)
	GetEntityRenderer(entityType reflect.Type) EntityRenderFunc
}
