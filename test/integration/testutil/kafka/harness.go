package kafka

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	kafkamodule "github.com/testcontainers/testcontainers-go/modules/kafka"
	redpandamodule "github.com/testcontainers/testcontainers-go/modules/redpanda"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	ConfluentLocalImage = "confluentinc/confluent-local:7.5.0"
	RedpandaImage       = "docker.redpanda.com/redpandadata/redpanda:v24.3.11"
)

type Provider string

const (
	ProviderKafka    Provider = "kafka"
	ProviderRedpanda Provider = "redpanda"
)

func RequireIntegration(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skip integration in short mode")
	}
	if v := os.Getenv("BIFROST_INTEGRATION"); v != "1" {
		t.Skip("set BIFROST_INTEGRATION=1 to run Docker integration tests")
	}
	testcontainers.SkipIfProviderIsNotHealthy(t)
}

func StartCluster(t *testing.T, ctx context.Context, provider Provider) []string {
	t.Helper()
	switch provider {
	case ProviderKafka:
		ctr, err := kafkamodule.Run(ctx, ConfluentLocalImage, kafkamodule.WithClusterID("bifrost-itest-kafka"))
		testcontainers.CleanupContainer(t, ctr)
		if err != nil {
			t.Fatalf("kafka run: %v", err)
		}
		brokers, err := ctr.Brokers(ctx)
		if err != nil {
			t.Fatalf("kafka brokers: %v", err)
		}
		return brokers
	case ProviderRedpanda:
		ctr, err := redpandamodule.Run(ctx, RedpandaImage, redpandamodule.WithAutoCreateTopics())
		testcontainers.CleanupContainer(t, ctr)
		if err != nil {
			t.Fatalf("redpanda run: %v", err)
		}
		broker, err := ctr.KafkaSeedBroker(ctx)
		if err != nil {
			t.Fatalf("redpanda seed: %v", err)
		}
		return []string{broker}
	default:
		t.Fatalf("unsupported provider: %s", provider)
		return nil
	}
}

func NewClient(t *testing.T, brokers []string, opts ...kgo.Opt) *kgo.Client {
	t.Helper()
	base := []kgo.Opt{kgo.SeedBrokers(brokers...)}
	cl, err := kgo.NewClient(append(base, opts...)...)
	if err != nil {
		t.Fatalf("new kafka client: %v", err)
	}
	return cl
}

func MustCreateTopic(t *testing.T, ctx context.Context, cl *kgo.Client, topic string, partitions int32) {
	t.Helper()
	adm := kadm.NewClient(cl)
	resp, err := adm.CreateTopics(ctx, partitions, 1, nil, topic)
	if err != nil {
		t.Fatalf("create topic %q: %v", topic, err)
	}
	for _, result := range resp.Sorted() {
		if result.Err != nil && !errors.Is(result.Err, kerr.TopicAlreadyExists) {
			t.Fatalf("create topic %q: %v", topic, result.Err)
		}
	}
}

func ProduceSync(t *testing.T, ctx context.Context, cl *kgo.Client, records ...*kgo.Record) {
	t.Helper()
	if err := cl.ProduceSync(ctx, records...).FirstErr(); err != nil {
		t.Fatalf("produce sync: %v", err)
	}
}

func WaitForRecords(ctx context.Context, cl *kgo.Client, topic string, count int, match func(*kgo.Record) bool) ([]*kgo.Record, error) {
	deadline := time.Now().Add(40 * time.Second)
	out := make([]*kgo.Record, 0, count)
	for len(out) < count && time.Now().Before(deadline) {
		fetches := cl.PollFetches(ctx)
		if err := fetches.Err(); err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, fmt.Errorf("poll: %w", err)
		}
		for _, r := range fetches.Records() {
			if r.Topic != topic {
				continue
			}
			if match == nil || match(r) {
				cp := *r
				out = append(out, &cp)
				if len(out) == count {
					break
				}
			}
		}
	}
	if len(out) < count {
		return nil, fmt.Errorf("timeout waiting for %d records on %s; got %d", count, topic, len(out))
	}
	return out, nil
}
