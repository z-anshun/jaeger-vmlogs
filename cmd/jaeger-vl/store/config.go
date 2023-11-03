package store

import "time"

const (
	defaultMaxSpanCount = int(1e7)
	defaultBatchSize    = 10_000
	defaultBatchDelay   = time.Second * 5

	defaultHttpListenAddr = ":9428"
	defaultMaxNumSpans    = 0
)

type Configuration struct {

	// Batch write size. Default is 10_000.
	BatchWriteSize int64 `yaml:"batch_write_size"`
	// Batch flush interval. Default is 5s.
	BatchFlushInterval time.Duration `yaml:"batch_flush_interval"`
	MaxSpanCount       int           `yaml:"max_span_count"`

	// The maximum number of spans to fetch per trace. If 0, no limits is set. Default 0.
	MaxNumSpans int `yaml:"max_num_spans"`

	// TCP address for http connections to listen vmlogs and explore metrics. Default localhost:9428
	HttpListenAddr string `yaml:"http_listen_addr"`
}

func (cfg *Configuration) SetDefaults() {
	if cfg.BatchWriteSize == 0 {
		cfg.BatchWriteSize = defaultBatchSize
	}
	if cfg.BatchFlushInterval == 0 {
		cfg.BatchFlushInterval = defaultBatchDelay
	}
	if cfg.MaxSpanCount == 0 {
		cfg.MaxSpanCount = defaultMaxSpanCount
	}

	if cfg.MaxNumSpans == 0 {
		cfg.MaxNumSpans = defaultMaxNumSpans
	}
	if len(cfg.HttpListenAddr) == 0 {
		cfg.HttpListenAddr = defaultHttpListenAddr
	}
}
