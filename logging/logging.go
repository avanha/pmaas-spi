// Package logging is how a plugin gets a logger, and how the levels of the server's loggers are looked at and changed.
//
// Every logger belongs to a component, named by dotted path, such as "plugin.esphome" or "core.eventmanager". A component's
// level is the one set for it, or else the nearest set for a component above it ("plugin" is above "plugin.esphome"), or else
// the default, which is Warn until someone says otherwise. A message below the level is dropped before it costs anything.
// What is set is kept by the server across restarts.
//
// A plugin asks for its logger when it is initialised, and logs through it:
//
//	p.log = logging.For(container)
//	p.log.Debug("poll finished", "device", name, "took", took)
//
// For a plugin with parts that deserve a level of their own, a part asks with a name: logging.For(container, "broker").
//
// Levels are the standard library's: Debug for what is only wanted when looking into something, Info for what happens in
// normal running, Warn for what is wrong but carried on from, and Error for what is not.
package logging

import (
	"log/slog"
	"sort"
	"strings"
)

// Component is one component's level.
type Component struct {
	// Name is the dotted name of the component, or empty for the default that applies to every component with no level of
	// its own.
	Name string

	// Level is the level in force for it.
	Level slog.Level

	// Set is whether the level was set for this component, as opposed to being inherited from a component above it or the
	// default. It is always true for the default.
	Set bool
}

// Controller looks at and changes the levels. It is safe to use from any goroutine.
type Controller interface {
	// Components lists the default and every component that has asked for a logger or has a level set, in name order.
	Components() []Component

	// SetLevel sets the level of a component, or of the default if the name is empty, and keeps it. It fails if the level is
	// not one of Debug, Info, Warn and Error, or the name is not a dotted name.
	SetLevel(name string, level slog.Level) error

	// ClearLevel takes away the level set for a component, which then inherits one. It does nothing to the default, which
	// can only be set.
	ClearLevel(name string) error
}

// Provider is what the server's container is, to the extent that it can hand out loggers and a Controller. A container that is
// not, such as a stand-in in a test, is dealt with by For and ControllerOf.
type Provider interface {
	// Logger returns the logger of the plugin's component, or of a part of it if names are given.
	Logger(names ...string) *slog.Logger

	// LogController returns the server's Controller.
	LogController() Controller
}

// For returns the logger a plugin is to use, given its container. With names, it is the logger of a part of the plugin: the
// component is the plugin's, then the names.
//
// A container that does not hand out loggers gets the process's default logger, so that a plugin can be tested with a stand-in
// for the container and still log.
func For(container any, names ...string) *slog.Logger {
	if p, ok := container.(Provider); ok {
		return p.Logger(names...)
	}

	return slog.Default().With("component", strings.Join(append([]string{"plugin"}, names...), "."))
}

// ControllerOf returns the Controller of a container, or nil if it does not have one.
func ControllerOf(container any) Controller {
	if p, ok := container.(Provider); ok {
		return p.LogController()
	}

	return nil
}

// Valid is whether the name is one a component can have: dotted, with no empty, blank or spaced part.
func Valid(name string) bool {
	if name == "" {
		return false
	}

	for _, part := range strings.Split(name, ".") {
		if part == "" || strings.ContainsAny(part, " \t\r\n=\"") {
			return false
		}
	}

	return true
}

// Depth is how far down the tree of components a name is: 0 for the default, 1 for "plugin", 2 for "plugin.esphome".
func Depth(name string) int {
	if name == "" {
		return 0
	}

	return strings.Count(name, ".") + 1
}

// SortByName puts components in the order they are listed in, the default first and each component before what is
// below it.
func SortByName(components []Component) {
	sort.Slice(components, func(i, j int) bool { return components[i].Name < components[j].Name })
}

// Levels are the levels that can be set, most detailed first.
var Levels = []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError}

// ValidLevel is whether the level is one of Levels.
func ValidLevel(level slog.Level) bool {
	for _, l := range Levels {
		if l == level {
			return true
		}
	}

	return false
}
