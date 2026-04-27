package artifacts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type TestArtifacts struct {
	Dir      string
	Stdout   string
	Stderr   string
	Config   string
	Metrics  string
	KeepOnOK bool
}

func New(t *testing.T, suite string) TestArtifacts {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".artifacts", suite, sanitize(t.Name())))
	if err != nil {
		t.Fatalf("artifacts abs path: %v", err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir artifacts: %v", err)
	}
	ta := TestArtifacts{
		Dir:     root,
		Stdout:  filepath.Join(root, "stdout.log"),
		Stderr:  filepath.Join(root, "stderr.log"),
		Config:  filepath.Join(root, "config.yaml"),
		Metrics: filepath.Join(root, "metrics.txt"),
	}
	t.Cleanup(func() {
		if t.Failed() || ta.KeepOnOK {
			return
		}
		_ = os.RemoveAll(root)
	})
	return ta
}

func (ta TestArtifacts) WriteConfig(data []byte) error {
	return os.WriteFile(ta.Config, data, 0o644)
}

func sanitize(s string) string {
	replacer := strings.NewReplacer("/", "_", " ", "_", ":", "_")
	return replacer.Replace(s)
}
