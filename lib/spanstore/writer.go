package spanstore

import (
	"context"
	"github.com/VictoriaMetrics/metrics"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/storage/spanstore"
)

var _ spanstore.Writer = (*SpanWriter)(nil)

type SpanWriter struct {
	logger hclog.Logger
	delay  time.Duration

	size  int64
	spans chan *model.Span

	finish chan struct{}
	done   sync.WaitGroup
}

func NewSpanWriter(logger hclog.Logger, size int64, delay time.Duration, maxSpanCount int) *SpanWriter {
	writer := &SpanWriter{
		logger: logger,
		delay:  delay,
		size:   size,
		spans:  make(chan *model.Span, size),
		finish: make(chan struct{}),
		done:   sync.WaitGroup{},
	}

	go writer.backgroundWriter(maxSpanCount)

	return writer

}

func (w *SpanWriter) backgroundWriter(maxSpanCount int) {
	pool := NewWorkerPool(w.logger, maxSpanCount)
	go pool.Work()

	batch := make([]*model.Span, 0, w.size)

	timer := time.After(w.delay)
	last := time.Now()

	for {
		w.done.Add(1)

		flush := false
		finish := false

		select {
		case span := <-w.spans:
			batch = append(batch, span)
			flush = len(batch) == cap(batch)
			if flush {
				w.logger.Debug("Flush due to batch size", "size", len(batch))
				numWritesWithBatchSize.Inc()
			}
		case <-timer:
			timer = time.After(w.delay)
			flush = time.Since(last) > w.delay && len(batch) > 0
			if flush {
				w.logger.Debug("Flush due to timer")
				numWritesWithFlushInterval.Inc()
			}
		case <-w.finish:
			finish = true
			flush = len(batch) > 0
			w.logger.Debug("Finish channel")
		}

		if flush {
			pool.WriteBatch(batch)

			batch = make([]*model.Span, 0, w.size)
			last = time.Now()
		}

		if finish {
			pool.Close()
		}
		// if finished will wait for the pool closed
		w.done.Done()
		if finish {
			break
		}
	}

}

func (w *SpanWriter) WriteSpan(ctx context.Context, span *model.Span) error {
	w.spans <- span
	return nil
}

func (w *SpanWriter) Close() error {
	w.logger.Debug("Waiting SpanWriter closed")
	w.finish <- struct{}{}
	w.done.Wait()
	w.logger.Debug("SpanWriter closed")
	return nil
}

var (
	numWritesWithBatchSize     = metrics.NewCounter("jaeger_vl_writes_with_batch_size_total")
	numWritesWithFlushInterval = metrics.NewCounter("jaeger_vl_writes_with_flush_interval_total")
)
