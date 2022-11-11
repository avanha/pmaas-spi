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
	GetTemplate(templateInfo *TemplateInfo) (CompiledTemplate, error)
	GetEntityRenderer(entityType reflect.Type) (EntityRenderer, error)
	RegisterEntityRenderer(entityType reflect.Type, renderFactory EntityRendererFactory)
	EnableStaticContent(staticContentDir string)
}
