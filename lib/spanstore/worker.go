package spanstore

import (
	"math"
	"sync"

	"github.com/VictoriaMetrics/metrics"
	"github.com/hashicorp/go-hclog"
	"github.com/jaegertracing/jaeger/model"
	"github.com/z-anshun/jaeger-vmlogs/app/vlstorage"
	"github.com/z-anshun/jaeger-vmlogs/lib/logstorage"
)

type WriteWorker struct {
	workerID   int32
	logger     hclog.Logger
	batch      []*model.Span
	workerDone chan *WriteWorker
	done       sync.WaitGroup

	lr *logstorage.LogRows
}

func (worker *WriteWorker) Work() {
	worker.done.Add(1)
	defer worker.done.Done()

	if err := worker.writeBatch(worker.batch); err != nil {
		worker.logger.Error("Could not write a batch of spans", "error", err, "worker_id", worker.workerID)
	}

	// everything is ok
	go worker.close() // try close
	return

	// need try?
}

func (worker *WriteWorker) Close() {
	worker.done.Wait()
}

func (worker *WriteWorker) close() {
	worker.workerDone <- worker
}

func (worker *WriteWorker) writeBatch(batch []*model.Span) error {
	parser := GetParser()
	defer PutParser(parser)
	for _, span := range batch {
		err := parser.ParseToTraceMsg(span)
		if err != nil {
			return err
		}
		worker.lr.MustAdd(logstorage.TenantID{}, span.StartTime.UnixNano(), parser.Fields)
		parser.reset()

		if worker.lr.NeedFlush() {
			vlstorage.MustAddRows(worker.lr)
		}
	}
	vlstorage.MustAddRows(worker.lr)
	worker.logger.Debug("Written spans", "size", len(batch))
	return nil
}

type WriteWorkerPool struct {
	logger  hclog.Logger
	finish  chan struct{}
	done    sync.WaitGroup
	batches chan []*model.Span

	maxSpanCount int
	mutex        sync.Mutex
	workers      workerHeap
	workerDone   chan *WriteWorker
}

func NewWorkerPool(logger hclog.Logger, maxSpanCount int) WriteWorkerPool {

	return WriteWorkerPool{
		logger:  logger,
		finish:  make(chan struct{}),
		done:    sync.WaitGroup{},
		batches: make(chan []*model.Span),

		maxSpanCount: maxSpanCount,

		mutex:      sync.Mutex{},
		workers:    newWorkerHeap(100),
		workerDone: make(chan *WriteWorker),
	}
}

func (pool *WriteWorkerPool) Work() {
	finish := false
	nexWorkerID := int32(1)

	for {

		pool.done.Add(1)
		select {
		case batch := <-pool.batches:
			batchSize := len(batch)
			if !pool.checkLimit(pendingSpanCount, batchSize) {
				numDiscardedSpans.Add(batchSize)
				pool.logger.Error("Discarding batch of spans due to exceeding pending span count", "batch_size", batchSize, "pending_span_count", pendingSpanCount, "max_span_count", pool.maxSpanCount)
			} else {
				worker := &WriteWorker{
					workerID:   nexWorkerID,
					logger:     pool.logger,
					batch:      batch,
					workerDone: pool.workerDone,
					done:       sync.WaitGroup{},
					lr:         logstorage.GetLogRows([]string{"service", "operationName"}, []string{}),
				}
				// [1,MaxInt32]
				nexWorkerID = nexWorkerID%math.MaxInt32 + 1

				pool.workers.AddWorker(worker)
				pendingSpanCount += batchSize
				go worker.Work()
			}

		case worker := <-pool.workerDone:
			pendingSpanCount -= len(worker.batch)
			if err := pool.workers.RemoveWorker(worker); err != nil {
				pool.logger.Error("could not remove worker", "worker", worker, "error", err)
			}
		case <-pool.finish:
			// close heap
			pool.workers.CloseWorkers()
			finish = true
		}
		pool.done.Done()

		if finish {
			break
		}
	}
}

func (pool *WriteWorkerPool) WriteBatch(batch []*model.Span) {
	pool.batches <- batch
}

func (pool *WriteWorkerPool) Close() {
	pool.finish <- struct{}{}
	pool.done.Wait()
}

func (pool *WriteWorkerPool) checkLimit(pendingSpanCount int, batchSize int) bool {
	if pool.maxSpanCount <= 0 {
		return true
	}

	// Check limit, add batchSize if within limit
	return pendingSpanCount+batchSize <= pool.maxSpanCount
}

var pendingSpanCount int

var (
	numDiscardedSpans = metrics.NewCounter("jaeger_vl_discarded_spans")
	_                 = metrics.NewGauge("jaeger_vl_pending_spans", func() float64 {
		return float64(pendingSpanCount)
	})
)
