package spanstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/storage/spanstore"
	"github.com/opentracing/opentracing-go"
	"github.com/z-anshun/jaeger-vmlogs/app/vlstorage"
	"github.com/z-anshun/jaeger-vmlogs/lib/logstorage"
)

var (
	errStartTimeRequired = errors.New("start time is required for search queries")
)

var _ spanstore.Reader = (*TraceReader)(nil)

const (
	minTimespanForProgressiveSearch       = time.Hour
	minTimespanForProgressiveSearchMargin = time.Minute
	maxProgressiveSteps                   = 4
	MaxMultiple                           = 10
)

var (
	defaultTenantIDs = []logstorage.TenantID{{0, 0}}
)

type TraceReader struct {
	logger      hclog.Logger
	maxNnmSpans int
}

func NewTraceReader(logger hclog.Logger, maxNumSpans int) *TraceReader {
	return &TraceReader{
		logger:      logger,
		maxNnmSpans: maxNumSpans,
	}
}

func (r *TraceReader) GetTrace(ctx context.Context, traceID model.TraceID) (*model.Trace, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetTrace")
	defer span.Finish()

	query := fmt.Sprintf("traceID: %s and _msg: *", traceID.String())

	span.SetTag("db.statement", query)
	span.SetTag("db.args", traceID.String())

	p, err := logstorage.ParseQuery(query)
	if err != nil {
		return nil, err
	}

	traces, err := r.getTraces(ctx, p, nil)
	if err != nil {
		return nil, err
	}

	if len(traces) == 0 {
		return nil, spanstore.ErrTraceNotFound
	}

	return traces[traceID], nil
}

func (r *TraceReader) getTraces(ctx context.Context, q *logstorage.Query, filter func(*model.Span) bool) (map[model.TraceID]*model.Trace, error) {

	traces := map[model.TraceID]*model.Trace{}
	lock := sync.Mutex{}
	var err error

	vlstorage.RunQuery(defaultTenantIDs, q, ctx.Done(), func(columns []logstorage.BlockColumn) {
		for _, column := range columns {
			if column.Name == "_msg" {
				lock.Lock()
				for _, value := range column.Values {
					span := &model.Span{}
					err = json.Unmarshal([]byte(value), span)
					if err != nil {
						continue
					}
					if filter != nil && !filter(span) {
						continue
					}
					if _, ok := traces[span.TraceID]; !ok {
						traces[span.TraceID] = &model.Trace{}
					}

					traces[span.TraceID].Spans = append(traces[span.TraceID].Spans, span)
				}
				lock.Unlock()
			}
		}
	})

	if err != nil {
		r.logger.Warn("Unmarshal trace err:", err)
	}

	return traces, nil
}

func (r *TraceReader) GetServices(ctx context.Context) ([]string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetServices")
	defer span.Finish()

	query := "service: *"
	q, err := logstorage.ParseQuery(query)
	if err != nil {
		return nil, err
	}

	span.SetTag("db.statement", query)
	//span.SetTag("db.args", args)

	services := r.getStrings(ctx, q, "service")
	return services, nil
}

func (r *TraceReader) GetOperations(ctx context.Context, params spanstore.OperationQueryParameters) ([]spanstore.Operation, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetOperations")
	defer span.Finish()

	query := "operationName: *"
	if len(params.ServiceName) != 0 {
		query = fmt.Sprintf("service:\"%s\" and operationName: *", params.ServiceName)
	}

	span.SetTag("db.statement", query)
	span.SetTag("db.args", params.ServiceName)

	q, err := logstorage.ParseQuery(query)
	if err != nil {
		return nil, err
	}
	operationsStr := r.getStrings(ctx, q, "operationName")

	operations := make([]spanstore.Operation, len(operationsStr))

	for i := range operationsStr {
		operations[i] = spanstore.Operation{Name: operationsStr[i]}
	}
	return operations, nil
}
func (r *TraceReader) getStrings(ctx context.Context, q *logstorage.Query, columnName string) []string {
	resMap := map[string]struct{}{}
	var lock sync.RWMutex

	vlstorage.RunQuery(defaultTenantIDs, q, ctx.Done(), func(columns []logstorage.BlockColumn) {

		for _, column := range columns {
			if column.Name == columnName {
				for _, orgVal := range column.Values {
					value := make([]byte, len(orgVal))
					// orgVal is generated by bytesutil.ToUnsafeString which makes it unsafe
					copy(value, orgVal)

					lock.RLock()
					if _, ok := resMap[string(value)]; ok {
						lock.RUnlock()
						continue
					}
					lock.RUnlock()

					lock.Lock()
					if _, ok := resMap[string(value)]; ok {
						continue
					}
					resMap[string(value)] = struct{}{}
					lock.Unlock()
				}

			}
		}
	})

	strs := make([]string, 0, len(resMap))
	for str := range resMap {
		strs = append(strs, str)
	}

	ss := getStringsSorter(strs)
	sort.Sort(ss)
	putStringsSorter(ss)
	return strs
}

func (r *TraceReader) FindTraces(ctx context.Context, params *spanstore.TraceQueryParameters) ([]*model.Trace, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "FindTraces")
	defer span.Finish()

	traceIDs, err := r.FindTraceIDs(ctx, params)
	if err != nil {
		return nil, err
	}

	if len(traceIDs) == 0 {
		return nil, nil
	}

	traceIDStrings := make([]string, 0, len(traceIDs))
	for _, traceID := range traceIDs {
		traceIDStrings = append(traceIDStrings, traceID.String())
	}
	// traceID: (id1 or id2)  and _msg: *
	query := fmt.Sprintf("traceID: (%s) and _msg: *", strings.Join(traceIDStrings, " or "))

	span.SetTag("db.statement", query)

	p, err := logstorage.ParseQuery(query)
	if err != nil {
		return nil, err
	}

	tracesMap, err := r.getTraces(ctx, p, func(span *model.Span) bool {
		if params.DurationMin != 0 && span.Duration < params.DurationMin {
			return false
		}

		if params.DurationMax != 0 && span.Duration > params.DurationMax {
			return false
		}
		return true
	})
	if err != nil {
		return nil, err
	}

	traces := make([]*model.Trace, 0, len(tracesMap))
	for _, trace := range tracesMap {
		traces = append(traces, trace)
	}
	sort.Slice(traces, func(i, j int) bool {
		if len(traces[i].Spans) == 0 {
			return false
		}
		if len(traces[j].Spans) == 0 {
			return true
		}
		return traces[i].Spans[0].StartTime.After(traces[j].Spans[0].StartTime)
	})

	if r.maxNnmSpans != 0 && len(traces) > r.maxNnmSpans {
		return traces[:r.maxNnmSpans], nil
	}

	// more
	if len(traces) > params.NumTraces*MaxMultiple {
		return traces[:params.NumTraces*MaxMultiple], nil
	}

	return traces, nil
}

func (r *TraceReader) FindTraceIDs(ctx context.Context, params *spanstore.TraceQueryParameters) ([]model.TraceID, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "FindTraceIDs")
	defer span.Finish()

	if params.StartTimeMin.IsZero() {
		return nil, errStartTimeRequired
	}

	end := params.StartTimeMax
	if end.IsZero() {
		end = time.Now()
	}

	fullTimeSpan := end.Sub(params.StartTimeMin)

	if fullTimeSpan < minTimespanForProgressiveSearch+minTimespanForProgressiveSearchMargin {
		return r.findTraceIDsInRange(ctx, params, params.StartTimeMin, end, map[model.TraceID]struct{}{})
	}

	timeSpan := fullTimeSpan
	for step := 0; step < maxProgressiveSteps; step++ {
		timeSpan /= 2
	}

	if timeSpan < minTimespanForProgressiveSearch {
		timeSpan = minTimespanForProgressiveSearch
	}

	found := make([]model.TraceID, 0)
	skip := map[model.TraceID]struct{}{}

	for step := 0; step < maxProgressiveSteps; step++ {
		if len(found) >= params.NumTraces {
			break
		}

		// last step
		if step == maxProgressiveSteps-1 {
			timeSpan = fullTimeSpan
		}

		start := end.Add(-timeSpan)
		if start.Before(params.StartTimeMin) {
			start = params.StartTimeMin
		}

		if start.After(end) {
			break
		}

		foundInRange, err := r.findTraceIDsInRange(ctx, params, start, end, skip)
		if err != nil {
			return nil, err
		}

		found = append(found, foundInRange...)

		end = start
		timeSpan *= 2
	}
	return found, nil
}

func (r *TraceReader) findTraceIDsInRange(ctx context.Context, params *spanstore.TraceQueryParameters, start, end time.Time, skip map[model.TraceID]struct{}) ([]model.TraceID, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "findTraceIDsInRange")
	defer span.Finish()

	if end.Before(start) || end == start {
		return []model.TraceID{}, nil
	}

	query := fmt.Sprintf("service: \"%s\" and _time: [%d,%d]", params.ServiceName, start.UnixMilli(), end.UnixMilli())

	if len(params.OperationName) != 0 {
		query += fmt.Sprintf(" and operationName: \"%s\"", params.OperationName)
	}

	for name, val := range params.Tags {
		query += fmt.Sprintf(" and %s:\"%s\"", name, val)
	}

	query += " and traceID: *"

	span.SetTag("db.statement", query)

	r.logger.Debug("will query:", query)

	q, err := logstorage.ParseQuery(query)
	if err != nil {
		return nil, err
	}

	traceIDs := []model.TraceID{} // TODO: sort by time
	lock := sync.RWMutex{}
	vlstorage.RunQuery(defaultTenantIDs, q, ctx.Done(), func(columns []logstorage.BlockColumn) {
		for _, column := range columns {
			if column.Name == "traceID" {
				for _, orgVal := range column.Values {
					value := make([]byte, len(orgVal))
					copy(value, orgVal)
					traceID, err := model.TraceIDFromString(string(value))
					if err != nil {
						// TODO: record
						continue
					}
					lock.RLock()
					if _, ok := skip[traceID]; ok {
						lock.RUnlock()
						continue
					}
					lock.RUnlock()

					lock.Lock()
					if _, ok := skip[traceID]; ok {
						continue
					}
					skip[traceID] = struct{}{}
					traceIDs = append(traceIDs, traceID)
					lock.Unlock()
				}
			}

		}

	})

	return traceIDs, nil
}

type stringsSorter struct {
	a []string
}

func (ss *stringsSorter) Len() int {
	return len(ss.a)
}
func (ss *stringsSorter) Swap(i, j int) {
	a := ss.a
	a[i], a[j] = a[j], a[i]
}
func (ss *stringsSorter) Less(i, j int) bool {
	a := ss.a
	return a[i] < a[j]
}

func getStringsSorter(a []string) *stringsSorter {
	v := stringsSorterPool.Get()
	if v == nil {
		return &stringsSorter{
			a: a,
		}
	}
	ss := v.(*stringsSorter)
	ss.a = a
	return ss
}

func putStringsSorter(ss *stringsSorter) {
	ss.a = nil
	stringsSorterPool.Put(ss)
}

var stringsSorterPool sync.Pool