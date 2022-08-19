package spi

import "net/http"

type IPMAASContainer interface {
	AddRoute(path string, handlerFunc http.HandlerFunc)
}
