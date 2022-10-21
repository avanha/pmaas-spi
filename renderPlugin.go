package spi

import (
	"net/http"
)

type IPMAASRenderPlugin interface {
	IPMAASPlugin
	RenderList(w http.ResponseWriter, r *http.Request, options RenderListOptions, items []interface{})
}
