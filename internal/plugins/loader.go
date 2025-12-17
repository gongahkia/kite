package plugins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"plugin"

	"github.com/rs/zerolog/log"
)

// Loader handles loading plugins from disk
type Loader struct {
	registry *Registry
	pluginDir string
}

// NewLoader creates a new plugin loader
func NewLoader(registry *Registry, pluginDir string) *Loader {
	return &Loader{
		registry:  registry,
		pluginDir: pluginDir,
	}
}

// LoadAll loads all plugins from the plugin directory
func (l *Loader) LoadAll(ctx context.Context) error {
	// Check if plugin directory exists
	if _, err := os.Stat(l.pluginDir); os.IsNotExist(err) {
		log.Warn().Str("dir", l.pluginDir).Msg("Plugin directory does not exist, skipping plugin loading")
		return nil
	}

	// Find all .so files
	files, err := filepath.Glob(filepath.Join(l.pluginDir, "*.so"))
	if err != nil {
		return fmt.Errorf("failed to list plugin files: %w", err)
	}

	if len(files) == 0 {
		log.Info().Str("dir", l.pluginDir).Msg("No plugins found")
		return nil
	}

	log.Info().Int("count", len(files)).Msg("Loading plugins")

	// Load each plugin
	loadedCount := 0
	for _, file := range files {
		if err := l.Load(ctx, file); err != nil {
			log.Error().Err(err).Str("file", file).Msg("Failed to load plugin")
			continue
		}
		loadedCount++
	}

	log.Info().Int("loaded", loadedCount).Int("total", len(files)).Msg("Plugin loading complete")

	return nil
}

// Load loads a single plugin from a file
func (l *Loader) Load(ctx context.Context, path string) error {
	log.Info().Str("path", path).Msg("Loading plugin")

	// Open the plugin
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin: %w", err)
	}

	// Look for New function
	newFunc, err := p.Lookup("New")
	if err != nil {
		return fmt.Errorf("plugin missing New function: %w", err)
	}

	// Call New function to get plugin instance
	newPluginFunc, ok := newFunc.(func() Plugin)
	if !ok {
		return fmt.Errorf("New function has incorrect signature")
	}

	plugin := newPluginFunc()

	// Initialize plugin
	if err := plugin.Init(make(map[string]interface{})); err != nil {
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	// Register plugin
	if err := l.registry.Register(plugin); err != nil {
		return fmt.Errorf("failed to register plugin: %w", err)
	}

	// Start plugin
	if err := plugin.Start(ctx); err != nil {
		return fmt.Errorf("failed to start plugin: %w", err)
	}

	log.Info().
		Str("plugin", plugin.Name()).
		Str("version", plugin.Version()).
		Msg("Plugin loaded successfully")

	return nil
}

// UnloadAll stops and unloads all plugins
func (l *Loader) UnloadAll(ctx context.Context) error {
	log.Info().Msg("Unloading all plugins")

	plugins := l.registry.List()

	for _, p := range plugins {
		if err := p.Stop(ctx); err != nil {
			log.Error().Err(err).Str("plugin", p.Name()).Msg("Failed to stop plugin")
		}

		if err := l.registry.Unregister(p.Name()); err != nil {
			log.Error().Err(err).Str("plugin", p.Name()).Msg("Failed to unregister plugin")
		}
	}

	log.Info().Int("count", len(plugins)).Msg("Plugins unloaded")

	return nil
}
