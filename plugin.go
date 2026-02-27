package spi

// IPMAASPlugin is the interface all plugins must implement.
// The core will call Init, Start, and Stop in sequence, all from the same go routine.
//
// All of these functions will be called on the plugin's main goroutine.
//
// The Stop function must not block.  If there is blocking work to be done, do the work in a new or separate
// goroutine, and inform the server that it's complete by closing the returned channel. The Stop function can also request additional
// callbacks by sending functions to the returned channel.
type IPMAASPlugin interface {
	Init(container IPMAASContainer)
	Start()
	Stop() chan func()
}
