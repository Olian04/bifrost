// Package config loads and validates bridge YAML configuration.
//
// Parse and Load return an error when required fields are missing or invalid. applyDefaults fills
// sensible defaults (logging level/format/stream, metrics listen address when metrics are enabled).
// MustParse and MustLoad panic instead of returning an error—use only when startup must abort on bad config.
package config

// Config is the top-level configuration for bifrost.
type Config struct {
	Clusters map[string]Cluster `yaml:"clusters"`
	Bridges  []Bridge           `yaml:"bridges"`
	Metrics  Metrics            `yaml:"metrics"`
	Logging  Logging            `yaml:"logging"`
}
