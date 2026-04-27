package scenario

import (
	"bytes"
	"text/template"
)

type ConfigData struct {
	FromTopic      string
	ToTopic        string
	FromCluster    string
	ToCluster      string
	BrokersFrom    []string
	BrokersTo      []string
	MetricsAddr    string
	BridgeName     string
	ConsumerGroup  string
	BatchSize      int
	Replicas       int
	OverrideKey    string
	HasOverrideKey bool
	ExtraHeaders   map[string]string
}

func RenderConfig(tmpl string, data ConfigData) ([]byte, error) {
	t, err := template.New("config").Parse(tmpl)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
