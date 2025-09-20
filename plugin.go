package spi

// IPMAASPlugin is the interface all plugins must implement.
// The core will call Init, Start and Stop in sequence, all from the same go routine.
type IPMAASPlugin interface {
	Init(container IPMAASContainer)
	Start()
	Stop()
}
