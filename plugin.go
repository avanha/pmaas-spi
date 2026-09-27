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
	// ShortName returns the plugin's stable, URL-safe identifier (e.g. "nestthermostat").
	// The server uses it to namespace every route the plugin registers via AddRoute/
	// AddRouteWithOptions/AddJsonRoute under "/plugins/<ShortName>/", and to derive the
	// plugin's default menu entry. It must be constant for the lifetime of the plugin
	// and must not change between releases, since it becomes part of the plugin's public
	// URL space.
	ShortName() string

	Init(container IPMAASContainer)
	Start()
	Stop() chan func()
}
