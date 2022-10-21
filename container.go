package spi

import (
	"net/http"
	"reflect"
)

type RenderListOptions struct {
	Title string
}

type IPMAASContainer interface {
	AddRoute(path string, handlerFunc http.HandlerFunc)
	RenderList(w http.ResponseWriter, r *http.Request, options RenderListOptions, items []interface{})
	GetTemplate(templateInfo *TemplateInfo) (ITemplate, error)
	GetEntityRenderer(entityType reflect.Type) (EntityRenderFunc, error)
	RegisterEntityRenderer(entityType reflect.Type, renderFactory EntityRendererFactory)
	RegisterStreamingEntityRenderer(entityType reflect.Type, streamingRenderFactory StreamingEntityRendererFactory)
}
