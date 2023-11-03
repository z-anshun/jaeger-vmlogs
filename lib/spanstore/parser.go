package spanstore

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/bytesutil"
	"github.com/jaegertracing/jaeger/model"
	"github.com/z-anshun/jaeger-vmlogs/lib/logstorage"
)

type Parser struct {
	Fields []logstorage.Field

	buf        []byte
	uniqueTags map[string][]string
}

func (p *Parser) reset() {
	fields := p.Fields
	for i := range fields {
		lf := &fields[i]
		lf.Name = ""
		lf.Value = ""
	}
	p.Fields = fields[:0]

	for key := range p.uniqueTags {
		delete(p.uniqueTags, key)
	}
	p.buf = p.buf[:0]
}

func (p *Parser) ParseToTraceMsg(span *model.Span) error {

	f := func(tags []model.KeyValue) {
		for k := range tags {
			tag := tags[k]
			p.uniqueTags[tag.GetKey()] = append(p.uniqueTags[tag.GetKey()], tag.AsString())
		}
	}

	// get keys
	f(span.Tags)

	if span.Process != nil {
		f(span.Process.Tags)
	}

	for i := range span.Logs {
		f(span.Logs[i].Fields)
	}

	for key := range p.uniqueTags {
		p.Fields, p.buf = appendTraceField(p.Fields, p.buf, key, strings.Join(p.uniqueTags[key], ","))
	}

	body, err := json.Marshal(span)
	if err != nil {
		return err
	}
	p.Fields, p.buf = appendTraceField(p.Fields, p.buf, "_msg", bytesutil.ToUnsafeString(body))
	p.Fields, p.buf = appendTraceField(p.Fields, p.buf, "service", span.Process.ServiceName)
	p.Fields, p.buf = appendTraceField(p.Fields, p.buf, "operationName", span.OperationName)
	p.Fields, p.buf = appendTraceField(p.Fields, p.buf, "traceID", span.TraceID.String())
	return nil
}

func appendTraceField(dst []logstorage.Field, dstBuf []byte, k string, value string) ([]logstorage.Field, []byte) {
	dstBuf = append(dstBuf, k...)

	dst = append(dst, logstorage.Field{
		Name:  k,
		Value: value,
	})
	return dst, dstBuf
}

func GetParser() *Parser {
	v := parserPool.Get()
	if v == nil {
		return &Parser{
			uniqueTags: map[string][]string{},
		}
	}
	return v.(*Parser)
}

func PutParser(p *Parser) {
	p.reset()
	parserPool.Put(p)
}

var parserPool sync.Pool
