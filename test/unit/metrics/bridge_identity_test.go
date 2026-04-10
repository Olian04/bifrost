package metrics_test

import (
	"testing"

	"github.com/lolocompany/bifrost/pkg/bridge"
	bifrostconfig "github.com/lolocompany/bifrost/pkg/config"
)

func TestBridgeIdentityLabelCountMatchesVec(t *testing.T) {
	id := bridge.IdentityFrom(bifrostconfig.Bridge{
		Name: "b1",
		From: bifrostconfig.BridgeTarget{
			Cluster: "src",
			Topic:   "t.in",
		},
		To: bifrostconfig.BridgeTarget{
			Cluster: "dst",
			Topic:   "t.out",
		},
	})
	if len(bridge.LabelNames) != len(id.LabelValues()) {
		t.Fatalf("LabelNames has %d names but LabelValues() has %d values", len(bridge.LabelNames), len(id.LabelValues()))
	}
}
