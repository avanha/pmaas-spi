package spi

type IPMAASPlugin interface {
	Init(container IPMAASContainer)
	Start()
	Stop()
}
