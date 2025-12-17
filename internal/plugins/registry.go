package plugins

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
)

// Registry manages all loaded plugins
type Registry struct {
	plugins map[string]Plugin
	mu      sync.RWMutex

	// Type-specific registries
	scrapers    map[string]ScraperPlugin
	processors  []ProcessorPlugin
	validators  []ValidatorPlugin
	exporters   map[string]ExporterPlugin
	middlewares []MiddlewarePlugin
}

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		plugins:     make(map[string]Plugin),
		scrapers:    make(map[string]ScraperPlugin),
		processors:  make([]ProcessorPlugin, 0),
		validators:  make([]ValidatorPlugin, 0),
		exporters:   make(map[string]ExporterPlugin),
		middlewares: make([]MiddlewarePlugin, 0),
	}
}

// Register registers a plugin
func (r *Registry) Register(plugin Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := plugin.Name()

	// Check if plugin already registered
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin already registered: %s", name)
	}

	// Register in main registry
	r.plugins[name] = plugin

	// Register in type-specific registry
	switch p := plugin.(type) {
	case ScraperPlugin:
		r.scrapers[name] = p
		log.Info().Str("plugin", name).Str("type", "scraper").Msg("Registered scraper plugin")

	case ProcessorPlugin:
		r.processors = append(r.processors, p)
		r.sortProcessors()
		log.Info().Str("plugin", name).Str("type", "processor").Msg("Registered processor plugin")

	case ValidatorPlugin:
		r.validators = append(r.validators, p)
		log.Info().Str("plugin", name).Str("type", "validator").Msg("Registered validator plugin")

	case ExporterPlugin:
		r.exporters[name] = p
		log.Info().Str("plugin", name).Str("type", "exporter").Msg("Registered exporter plugin")

	case MiddlewarePlugin:
		r.middlewares = append(r.middlewares, p)
		r.sortMiddlewares()
		log.Info().Str("plugin", name).Str("type", "middleware").Msg("Registered middleware plugin")

	default:
		log.Info().Str("plugin", name).Str("type", "generic").Msg("Registered generic plugin")
	}

	return nil
}

// Unregister unregisters a plugin
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	plugin, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	// Remove from type-specific registry
	switch p := plugin.(type) {
	case ScraperPlugin:
		delete(r.scrapers, name)

	case ProcessorPlugin:
		for i, proc := range r.processors {
			if proc.Name() == name {
				r.processors = append(r.processors[:i], r.processors[i+1:]...)
				break
			}
		}

	case ValidatorPlugin:
		for i, val := range r.validators {
			if val.Name() == name {
				r.validators = append(r.validators[:i], r.validators[i+1:]...)
				break
			}
		}

	case ExporterPlugin:
		delete(r.exporters, name)

	case MiddlewarePlugin:
		for i, mw := range r.middlewares {
			if mw.Name() == name {
				r.middlewares = append(r.middlewares[:i], r.middlewares[i+1:]...)
				break
			}
		}
	}

	// Remove from main registry
	delete(r.plugins, name)

	log.Info().Str("plugin", name).Msg("Unregistered plugin")
	return nil
}

// Get retrieves a plugin by name
func (r *Registry) Get(name string) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}

	return plugin, nil
}

// GetScraper retrieves a scraper plugin by name
func (r *Registry) GetScraper(name string) (ScraperPlugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	scraper, exists := r.scrapers[name]
	if !exists {
		return nil, fmt.Errorf("scraper plugin not found: %s", name)
	}

	return scraper, nil
}

// GetExporter retrieves an exporter plugin by name
func (r *Registry) GetExporter(name string) (ExporterPlugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	exporter, exists := r.exporters[name]
	if !exists {
		return nil, fmt.Errorf("exporter plugin not found: %s", name)
	}

	return exporter, nil
}

// ListScrapers returns all registered scraper plugins
func (r *Registry) ListScrapers() []ScraperPlugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	scrapers := make([]ScraperPlugin, 0, len(r.scrapers))
	for _, s := range r.scrapers {
		scrapers = append(scrapers, s)
	}

	return scrapers
}

// ListProcessors returns all registered processor plugins (sorted by priority)
func (r *Registry) ListProcessors() []ProcessorPlugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return append([]ProcessorPlugin{}, r.processors...)
}

// ListValidators returns all registered validator plugins
func (r *Registry) ListValidators() []ValidatorPlugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return append([]ValidatorPlugin{}, r.validators...)
}

// ListExporters returns all registered exporter plugins
func (r *Registry) ListExporters() []ExporterPlugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	exporters := make([]ExporterPlugin, 0, len(r.exporters))
	for _, e := range r.exporters {
		exporters = append(exporters, e)
	}

	return exporters
}

// ListMiddlewares returns all registered middleware plugins (sorted by order)
func (r *Registry) ListMiddlewares() []MiddlewarePlugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return append([]MiddlewarePlugin{}, r.middlewares...)
}

// List returns all registered plugins
func (r *Registry) List() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		plugins = append(plugins, p)
	}

	return plugins
}

// Count returns the total number of registered plugins
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.plugins)
}

// sortProcessors sorts processors by priority (descending)
func (r *Registry) sortProcessors() {
	for i := 0; i < len(r.processors)-1; i++ {
		for j := i + 1; j < len(r.processors); j++ {
			if r.processors[i].Priority() < r.processors[j].Priority() {
				r.processors[i], r.processors[j] = r.processors[j], r.processors[i]
			}
		}
	}
}

// sortMiddlewares sorts middlewares by order (ascending)
func (r *Registry) sortMiddlewares() {
	for i := 0; i < len(r.middlewares)-1; i++ {
		for j := i + 1; j < len(r.middlewares); j++ {
			if r.middlewares[i].Order() > r.middlewares[j].Order() {
				r.middlewares[i], r.middlewares[j] = r.middlewares[j], r.middlewares[i]
			}
		}
	}
}
