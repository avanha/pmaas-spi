package spi

import (
	"net/http"
	"time"
)

// PluginVersion identifies one running plugin's build version, for status-page display.
// Version comes from the running binary's own embedded module build info (see
// runtime/debug.ReadBuildInfo): a semantic version tag if the plugin's module was fetched at
// one, a pseudo-version encoding its git commit hash and commit time if not, or "(devel)" when
// built from a local replace/workspace directory rather than a fixed module version - as this
// project's own development builds normally are, via its go.work file. Empty if no build info
// is available at all.
type PluginVersion struct {
	ShortName string
	Version   string
}

// LoadAverage is the host system's 1/5/15-minute load average. See ServerStatus.LoadAverage
// for when it's available.
type LoadAverage struct {
	Load1  float64
	Load5  float64
	Load15 float64
}

// MemoryStats is this server process's own memory usage (from runtime.MemStats) - not the
// host system's total memory, which would need platform-specific code to obtain portably.
type MemoryStats struct {
	AllocBytes      uint64
	TotalAllocBytes uint64
	SysBytes        uint64
	NumGC           uint32
}

// ServerStatus is the data behind the server's root ("/") status page - see
// IPMAASContainer.ProvideRootStatusHandler.
type ServerStatus struct {
	Uptime  time.Duration
	Plugins []PluginVersion

	// LoadAverage is nil where a load average isn't available - currently anywhere other than
	// Linux (this project's actual deployment target), or if it couldn't be read for any
	// other reason.
	LoadAverage *LoadAverage

	Memory MemoryStats
}

// RootStatusHandlerFunc handles a request to the server's root ("/") path, given a freshly
// computed ServerStatus - see IPMAASContainer.ProvideRootStatusHandler.
type RootStatusHandlerFunc func(w http.ResponseWriter, r *http.Request, status ServerStatus)
