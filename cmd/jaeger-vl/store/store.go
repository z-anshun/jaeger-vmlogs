package store

import (
	"io"
	"time"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/logger"
	"github.com/hashicorp/go-hclog"
	"github.com/jaegertracing/jaeger/plugin/storage/grpc/shared"
	"github.com/jaegertracing/jaeger/storage/dependencystore"
	"github.com/jaegertracing/jaeger/storage/spanstore"
	"github.com/z-anshun/jaeger-vmlogs/app/vlselect"
	"github.com/z-anshun/jaeger-vmlogs/app/vlstorage"
	vlspanstorage "github.com/z-anshun/jaeger-vmlogs/lib/spanstore"
)

var (
	_ shared.StoragePlugin = (*Store)(nil)
	//_ shared.ArchiveStoragePlugin      = (*Store)(nil)
	//_ shared.StreamingSpanWriterPlugin = (*Store)(nil)
	_ io.Closer = (*Store)(nil)
)

type Store struct {
	writer spanstore.Writer
	reader spanstore.Reader
}

func NewStore(logger hclog.Logger, conf *Configuration) *Store {
	return &Store{
		writer: vlspanstorage.NewSpanWriter(logger,
			conf.BatchWriteSize, conf.BatchFlushInterval, conf.MaxSpanCount),
		reader: vlspanstorage.NewTraceReader(logger, conf.MaxNumSpans),
	}
}

func (s *Store) DependencyReader() dependencystore.Reader {
	return vlspanstorage.NewDependencyStore()
}

func (s *Store) SpanReader() spanstore.Reader {
	return s.reader
}

func (s *Store) SpanWriter() spanstore.Writer {
	return s.writer
}

func (s *Store) Close() error {
	startTime := time.Now()

	vlstorage.Stop()
	vlselect.Stop()
	logger.Infof("the VictoriaLogs has been stopped in %.3f seconds", time.Since(startTime).Seconds())
	return nil
}
