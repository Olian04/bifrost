package run

import (
	"log/slog"
	"strings"

	"github.com/lolocompany/bifrost/pkg/config"
)

// logKafkaClientDebug logs one structured snapshot of cluster client settings at debug level.
// clusterName is the config clusters map key; clientRole is "producer" or "consumer".
// Does not log secrets (SASL passwords) or full broker hostnames—only counts and tuning fields.
func logKafkaClientDebug(clusterName, clientRole string, env *config.Cluster) {
	if env == nil {
		return
	}
	attrs := []any{
		"cluster", clusterName,
		"client_role", clientRole,
		"seed_broker_count", len(env.Brokers),
		"tls_enabled", env.TLS.Enabled,
	}
	mech := strings.ToLower(strings.TrimSpace(env.SASL.Mechanism))
	if mech == "" {
		mech = "none"
	}
	attrs = append(attrs, "sasl_mechanism", mech)
	if id := strings.TrimSpace(env.Client.ClientID); id != "" {
		attrs = append(attrs, "client_id", id)
	}
	if d := strings.TrimSpace(env.Client.DialTimeout); d != "" {
		attrs = append(attrs, "client_dial_timeout", d)
	}
	if ro := strings.TrimSpace(env.Client.RequestTimeoutOverhead); ro != "" {
		attrs = append(attrs, "client_request_timeout_overhead", ro)
	}

	switch clientRole {
	case "consumer":
		il := strings.ToLower(strings.TrimSpace(env.Consumer.IsolationLevel))
		if il == "" {
			il = "read_uncommitted"
		}
		attrs = append(attrs, "consumer_isolation_level", il)
		if env.Consumer.FetchMaxBytes != nil {
			attrs = append(attrs, "consumer_fetch_max_bytes", *env.Consumer.FetchMaxBytes)
		}
		if env.Consumer.FetchMaxPartitionBytes != nil {
			attrs = append(attrs, "consumer_fetch_max_partition_bytes", *env.Consumer.FetchMaxPartitionBytes)
		}
		if w := strings.TrimSpace(env.Consumer.FetchMaxWait); w != "" {
			attrs = append(attrs, "consumer_fetch_max_wait", w)
		}
	case "producer":
		acks := strings.ToLower(strings.TrimSpace(env.Producer.RequiredAcks))
		if acks == "" {
			acks = "all"
		}
		attrs = append(attrs, "producer_required_acks", acks)
		if c := strings.TrimSpace(env.Producer.BatchCompression); c != "" {
			attrs = append(attrs, "producer_batch_compression", c)
		}
		if env.Producer.BatchMaxBytes != nil {
			attrs = append(attrs, "producer_batch_max_bytes", *env.Producer.BatchMaxBytes)
		}
		if l := strings.TrimSpace(env.Producer.Linger); l != "" {
			attrs = append(attrs, "producer_linger", l)
		}
	}

	slog.Debug("kafka client config", attrs...)
}
