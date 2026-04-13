package version_test

import (
	"testing"

	"github.com/lolocompany/bifrost/pkg/version"
)

func TestInfo_StableShape(t *testing.T) {
	info := version.Info()
	if info.Version == "" {
		t.Fatal("Version must be non-empty (at least unknown)")
	}
	if info.Revision == "" {
		t.Fatal("Revision must be non-empty (at least unknown)")
	}
	if info.BuildTime == "" {
		t.Fatal("BuildTime must be non-empty (at least unknown)")
	}
}
