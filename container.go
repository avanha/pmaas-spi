package spi

import (
	"crypto/tls"
	"fmt"
	"io/fs"
	"net/http"
	"reflect"

	"github.com/avanha/pmaas-spi/entity"
	"github.com/avanha/pmaas-spi/events"
)

// PluginRoutePrefix is the fixed prefix under which every plugin's routes are namespaced.
// The server always constructs a plugin's full route pattern as PluginRoutePrefix + ShortName +
// "/" + the relative path given to AddRoute/AddRouteWithOptions/AddJsonRoute.
const PluginRoutePrefix = "/plugins/"

// PluginAssetPrefix is the fixed prefix under which a plugin's static content and template
// Scripts/Styles are served. It's deliberately disjoint from PluginRoutePrefix: static content is
// mounted as a whole subtree at a plugin's asset root, so if it shared PluginRoutePrefix, that
// mount would collide with the plugin's own list route (both would be the exact pattern
// "/plugins/<name>/", which net/http.ServeMux refuses to register twice) regardless of what
// relative paths the plugin chooses for AddRoute.
const PluginAssetPrefix = "/plugins-assets/"

type RenderListOptions struct {
	Title string

	// TitleSuffix is optional additional text shown after Title, rendered smaller/muted (see
	// layout.htmlt's "title-suffix" class) since it can be an arbitrarily long string that would
	// otherwise threaten to overflow the page header - e.g. the root status page's assembly
	// name/version (see pmaas-plugin-basicwebui's statusTitle). Ignored (no visual effect) when
	// empty.
	TitleSuffix string

	Header any
}

type HttpHandlerOptions struct {
	SupportsXsrfValidation bool
	RequiresXsrfValidation bool

	// IncludeInMenu controls whether this route gets a navigation menu entry. If nil, the
	// server applies its default: a plugin's list route (registered with relativePath "")
	// defaults to included, every other route defaults to excluded. Set explicitly to
	// override either default.
	IncludeInMenu *bool

	// MenuLabel is the display text for this route's menu entry. If empty, the server falls
	// back to the plugin's ShortName() for the list route, or the route's relative path for
	// any other route.
	MenuLabel string

	// MenuIcon is an optional icon identifier for this route's menu entry.
	MenuIcon string
}

// MenuEntry describes one navigation menu item, built from the routes plugins have registered
// as menu-visible. Only a plugin's list route (relativePath "") can have Children - the server
// supports a single level of nesting, so a non-list-route entry never has children of its own.
type MenuEntry struct {
	Label    string
	Icon     string
	Path     string
	Children []MenuEntry
}

type RequestObjectFactoryFunc func() any

type JsonHandlerFunc func(w http.ResponseWriter, r *http.Request, request any) (any, error)

// IPMAASContainer is an interface for plugins to interact with the PMAAS server.
type IPMAASContainer interface {
	// AddRoute registers a handler for a path relative to this plugin's namespace. The
	// server serves it at PluginFullPath(pluginShortName, path) - e.g. relativePath "" is
	// the plugin's list route ("/plugins/<name>/"), and relativePath "callback" is served at
	// "/plugins/<name>/callback". Plugins never include the "/plugins/<name>/" prefix
	// themselves.
	AddRoute(path string, handlerFunc http.HandlerFunc)
	// AddRouteWithOptions is AddRoute with explicit HttpHandlerOptions, e.g. to control this
	// route's menu visibility/label/icon.
	AddRouteWithOptions(path string, handlerFunc http.HandlerFunc, options *HttpHandlerOptions)
	// AddJsonRoute is AddRoute for a JSON request/response handler; path is relative exactly
	// as in AddRoute.
	AddJsonRoute(path string, requestFactoryFnc RequestObjectFactoryFunc, handlerFunc JsonHandlerFunc)
	// GetMenu returns the current navigation menu, built from every plugin's menu-visible
	// routes, in plugin-registration order.
	GetMenu() []MenuEntry
	// RouteFullPath returns the full, server-rooted path for one of this plugin's own routes at
	// the given path relative to its own namespace - equivalent to
	// PluginFullPath(ShortName(), relativePath), but without the plugin needing to know or pass
	// its own ShortName(). Useful for building a fully-qualified URL to one of the plugin's own
	// routes for use outside the server itself (e.g. an OAuth redirect URI).
	RouteFullPath(relativePath string) string
	// AssetFullPath returns the full, server-rooted path for one of this plugin's own static
	// content/template assets at the given path relative to its own asset root - equivalent to
	// PluginAssetFullPath(ShortName(), relativePath), but without the plugin needing to know or
	// pass its own ShortName().
	AssetFullPath(relativePath string) string
	BroadcastEvent(entityEventId string, event any) error
	RenderList(w http.ResponseWriter, r *http.Request, options RenderListOptions, items []interface{})
	GetTemplate(templateInfo *TemplateInfo) (CompiledTemplate, error)
	GetEntityRenderer(entityType reflect.Type) (EntityRenderer, error)
	RegisterEntityRenderer(entityType reflect.Type, renderFactory EntityRendererFactory)
	EnableStaticContent(staticContentDir string)

	// GetBaseUrl returns the base URL (scheme://host[:port]) that should be used to build absolute
	// URLs (e.g. OAuth redirect URIs) back to this server, chosen from the server's configured list
	// of base URLs by matching the given request's Host header. This lets a single deployment serve
	// requests under multiple externally-reachable names (e.g. "localhost" during development and a
	// real hostname in production) without trusting the request itself for scheme/host: the returned
	// value always comes from server configuration, never from request headers directly. Returns an
	// error if no configured base URL matches the request's Host header.
	GetBaseUrl(r *http.Request) (string, error)

	// ProvideContentFS Registers an io/fs.FS instance that the server can use to read plugin resources such as
	// templates or static file content for serving over HTTP.
	ProvideContentFS(fs fs.FS, prefix string)

	// RegisterEntity Registers an entity with server.  This gives it a unique name, which can be used for
	// further interaction with the entity.
	//
	// The server guarantees that the provided stubFactoryFn will only ever be called on the
	// plugin's main goroutine.
	RegisterEntity(
		uniqueData string,
		entityType reflect.Type,
		name string,
		stubFactoryFn EntityStubFactoryFunc) (string, error)

	// DeregisterEntity Removes an entity previously registered with the server.  Pass the ID returned from the
	// previous call to RegisterEntity.
	DeregisterEntity(id string) error

	// AssertEntityType Verifies that an entity with the given ID exists and is of the passed type
	AssertEntityType(pmaasEntityId string, entityType reflect.Type) error

	// GetEntities Returns a list of entities that match the supplied predicate.
	GetEntities(predicate func(info *entity.RegisteredEntityInfo) bool) ([]entity.RegisteredEntityInfo, error)

	// InvokeOnEntity Invokes the supplied function on the plugin-runner goroutine of the plugin that owns the specified
	//entity, supplying the entity.
	InvokeOnEntity(id string, function func(entity any)) error

	// RegisterEventReceiver Registers a receiver for events.  If successful, returns an integer handle that can be used
	// to deregister the handler in the future.
	RegisterEventReceiver(predicate events.EventPredicate, receiver events.EventReceiver) (int, error)

	// DeregisterEventReceiver Removes a previously registered event receiver
	DeregisterEventReceiver(receiverHandle int) error

	// EnqueueOnPluginGoRoutine Enqueues the passed function for execution on the plugin's main GoRoutine.
	// Returns an error if the function cannot be enqueued.  This method returns as soon as the function
	// is enqueued, it does not wait for execution.  If you need the results of the execution, you'll need to
	// orchestrate that in the function.  For example, the function can send the result back via a channel.
	// Warning: Calling this function when already executing on the plugin's main GoRoutine will result in deadlock.
	EnqueueOnPluginGoRoutine(f func()) error

	// EnqueueOnServerGoRoutine Enqueues the passed functions for execution on the server's main GoRoutine.
	// The server's main GoRoutine is the one used to call PMAAS.Run().
	// Use this to execute callbacks registered during the server configuration phase.
	EnqueueOnServerGoRoutine(invocations []func()) error

	// ClosedCallbackChannel returns an already closed callback channel.  You can use this when you know
	// there will not be a need for any callbacks.
	ClosedCallbackChannel() chan func()

	SaveConfig(config any) error
	LoadConfig(f func(typeName string) any) (any, error)

	// ProvideTLSCertificate registers this plugin as the server's TLS certificate provider. The
	// getCertificateFunc is stored as-is and handed to Go's net/http.Server as
	// tls.Config.GetCertificate, so it's invoked fresh on every incoming TLS handshake - a plugin
	// that renews/rotates its certificate in the background (e.g. an ACME client) doesn't need to
	// tell the server anything when that happens: the next handshake after a renewal automatically
	// picks up whatever certificate the function now returns. Because it's called concurrently, once
	// per handshake, getCertificateFunc must be safe for concurrent use.
	//
	// At most one plugin may call this - it returns an error if a certificate provider has already
	// been registered by another plugin. If no plugin ever calls it, the server continues to serve
	// plain HTTP exactly as it did before this method existed.
	//
	// Must be called during Init or Start, before this plugin's Start returns - the server reads the
	// registered provider (if any) only once, when it starts listening, which happens strictly after
	// every plugin has finished starting.
	ProvideTLSCertificate(getCertificateFunc func(*tls.ClientHelloInfo) (*tls.Certificate, error)) error

	// ProvideRootStatusHandler registers this plugin as the server's root ("/") status page
	// handler. handlerFunc is called for every request to "/", along with a freshly computed
	// ServerStatus - the plugin owns rendering it (e.g. via its own templates), keeping that
	// presentation logic out of pmaas-core.
	//
	// At most one plugin may call this - it returns an error if a handler has already been
	// registered by another plugin. If no plugin ever calls it, the server serves a minimal
	// placeholder page at "/" instead.
	//
	// Must be called during Init or Start, before this plugin's Start returns - the same
	// timing requirement as ProvideTLSCertificate, and for the same reason: the server only
	// reads the registered handler (if any) once, when it starts listening.
	ProvideRootStatusHandler(handlerFunc RootStatusHandlerFunc) error
}

// Exec enqueues f for execution on the plugin's own goroutine and blocks until it runs,
// returning its result. It's the "ask" half of the tell/ask split already used elsewhere in this
// codebase (see pmaas-common/mailbox.Mailbox.Exec) - EnqueueOnPluginGoRoutine alone is
// fire-and-forget (tell); Go doesn't support generic interface methods, so this has to be a
// package-level function taking the container as a parameter rather than a method on
// IPMAASContainer itself.
//
// The same reentrancy warning as EnqueueOnPluginGoRoutine applies: calling Exec from a function
// already executing on the plugin's own goroutine deadlocks, since nothing will ever run f.
func Exec[R any](container IPMAASContainer, f func() R) (R, error) {
	resultCh := make(chan R, 1)

	err := container.EnqueueOnPluginGoRoutine(func() {
		resultCh <- f()
	})

	if err != nil {
		var zero R
		return zero, fmt.Errorf("unable to enqueue function execution on plugin goroutine: %w", err)
	}

	return <-resultCh, nil
}

// ExecValueFunctionOnPluginGoRoutine is Exec with a caller-supplied fallback value for when
// enqueueing fails, and a message to wrap the resulting error with. Kept for existing callers;
// new code should prefer Exec directly, since swallowing the enqueue failure into defaultValueFn
// makes "enqueue failed" indistinguishable from "f legitimately returned this value".
func ExecValueFunctionOnPluginGoRoutine[R any](
	container IPMAASContainer,
	f func() R,
	defaultValueFn func() R,
	errorMessage string) (R, error) {
	result, err := Exec(container, f)

	if err != nil {
		return defaultValueFn(), fmt.Errorf("%s: %w", errorMessage, err)
	}

	return result, nil
}
