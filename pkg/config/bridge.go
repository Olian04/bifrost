package config

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// Bridge defines one directional relay between two clusters and topics (from → to).
type Bridge struct {
	Name          string       `yaml:"name"`
	From          BridgeTarget `yaml:"from"`
	To            BridgeTarget `yaml:"to"`
	ConsumerGroup string       `yaml:"consumer_group"`
}

// BridgeTarget references a cluster name (key in clusters) and a single topic.
type BridgeTarget struct {
	Cluster string `yaml:"cluster"`
	Topic   string `yaml:"topic"`
}

func (b *Bridge) validate(clusters map[string]Cluster) error {
	if strings.TrimSpace(b.Name) == "" {
		return errors.New("name is required")
	}
	if err := b.From.validate("from", clusters); err != nil {
		return err
	}
	if err := b.To.validate("to", clusters); err != nil {
		return err
	}
	if b.From.Cluster == b.To.Cluster && b.From.Topic == b.To.Topic {
		return errors.New("from and to cannot be the same cluster and topic")
	}
	return nil
}

func (t *BridgeTarget) validate(role string, clusters map[string]Cluster) error {
	if strings.TrimSpace(t.Cluster) == "" {
		return fmt.Errorf("%s.cluster is required", role)
	}
	if _, ok := clusters[t.Cluster]; !ok {
		return fmt.Errorf("%s.cluster %q is not defined under clusters", role, t.Cluster)
	}
	if strings.TrimSpace(t.Topic) == "" {
		return fmt.Errorf("%s.topic is required", role)
	}
	return nil
}

// EffectiveConsumerGroup returns the consumer group for this bridge.
func (b *Bridge) EffectiveConsumerGroup() string {
	if strings.TrimSpace(b.ConsumerGroup) != "" {
		return strings.TrimSpace(b.ConsumerGroup)
	}
	return "bifrost-" + sanitizeName(b.Name)
}

func sanitizeName(s string) string {
	repl := strings.NewReplacer(" ", "-", "_", "-")
	s = strings.ToLower(strings.TrimSpace(repl.Replace(s)))
	var b strings.Builder
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		case r == '-' || r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	out := strings.Trim(b.String(), "-.")
	for strings.Contains(out, "--") {
		out = strings.ReplaceAll(out, "--", "-")
	}
	if out == "" {
		return "default"
	}
	return out
}
