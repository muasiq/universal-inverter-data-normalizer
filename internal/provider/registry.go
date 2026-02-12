package provider

import (
	"fmt"
	"sync"
)

// Factory is a function that creates a new Provider instance.
type Factory func() Provider

// Registry manages the available provider implementations.
type Registry struct {
	mu        sync.RWMutex
	factories map[string]Factory
}

// NewRegistry creates a new empty provider registry.
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]Factory),
	}
}

// Register adds a provider factory to the registry.
func (r *Registry) Register(name string, factory Factory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[name] = factory
}

// Create instantiates a provider by name.
func (r *Registry) Create(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, ok := r.factories[name]
	if !ok {
		return nil, fmt.Errorf("unknown provider type: %q (registered: %v)", name, r.ListProviders())
	}
	return factory(), nil
}

// ListProviders returns the names of all registered providers.
func (r *Registry) ListProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

// DefaultRegistry is the global provider registry.
var DefaultRegistry = NewRegistry()

// Register adds a provider factory to the default registry.
func Register(name string, factory Factory) {
	DefaultRegistry.Register(name, factory)
}
