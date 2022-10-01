package spi

import (
	"io"
	"net/http"
)

type IPMAASRenderPlugin interface {
	IPMAASPlugin
	RenderList(w http.ResponseWriter, r *http.Request, items []interface{})
	RenderEntity(w io.Writer, item any)
}

