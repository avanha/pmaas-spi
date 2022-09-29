package spi

import "net/http"

type ITemplate interface {}

type IPMAASContainer interface {
	AddRoute(path string, handlerFunc http.HandlerFunc)
	RenderList(w http.ResponseWriter, r *http.Request, items []interface{})
	GetTemplate(path string) (ITemplate, error)
}
