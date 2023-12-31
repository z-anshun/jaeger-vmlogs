// Code generated by qtc from "query_response.qtpl". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

//line app/vlselect/logsql/query_response.qtpl:1
package logsql

//line app/vlselect/logsql/query_response.qtpl:1
import (
	"github.com/z-anshun/jaeger-vmlogs/lib/logstorage"
)

// JSONRow creates JSON row from the given fields.

//line app/vlselect/logsql/query_response.qtpl:8
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line app/vlselect/logsql/query_response.qtpl:8
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line app/vlselect/logsql/query_response.qtpl:8
func StreamJSONRow(qw422016 *qt422016.Writer, columns []logstorage.BlockColumn, rowIdx int) {
//line app/vlselect/logsql/query_response.qtpl:8
	qw422016.N().S(`{`)
//line app/vlselect/logsql/query_response.qtpl:10
	c := &columns[0]

//line app/vlselect/logsql/query_response.qtpl:11
	qw422016.N().Q(c.Name)
//line app/vlselect/logsql/query_response.qtpl:11
	qw422016.N().S(`:`)
//line app/vlselect/logsql/query_response.qtpl:11
	qw422016.N().Q(c.Values[rowIdx])
//line app/vlselect/logsql/query_response.qtpl:12
	columns = columns[1:]

//line app/vlselect/logsql/query_response.qtpl:13
	for colIdx := range columns {
//line app/vlselect/logsql/query_response.qtpl:14
		c := &columns[colIdx]

//line app/vlselect/logsql/query_response.qtpl:14
		qw422016.N().S(`,`)
//line app/vlselect/logsql/query_response.qtpl:15
		qw422016.N().Q(c.Name)
//line app/vlselect/logsql/query_response.qtpl:15
		qw422016.N().S(`:`)
//line app/vlselect/logsql/query_response.qtpl:15
		qw422016.N().Q(c.Values[rowIdx])
//line app/vlselect/logsql/query_response.qtpl:16
	}
//line app/vlselect/logsql/query_response.qtpl:16
	qw422016.N().S(`}`)
//line app/vlselect/logsql/query_response.qtpl:17
	qw422016.N().S(`
`)
//line app/vlselect/logsql/query_response.qtpl:18
}

//line app/vlselect/logsql/query_response.qtpl:18
func WriteJSONRow(qq422016 qtio422016.Writer, columns []logstorage.BlockColumn, rowIdx int) {
//line app/vlselect/logsql/query_response.qtpl:18
	qw422016 := qt422016.AcquireWriter(qq422016)
//line app/vlselect/logsql/query_response.qtpl:18
	StreamJSONRow(qw422016, columns, rowIdx)
//line app/vlselect/logsql/query_response.qtpl:18
	qt422016.ReleaseWriter(qw422016)
//line app/vlselect/logsql/query_response.qtpl:18
}

//line app/vlselect/logsql/query_response.qtpl:18
func JSONRow(columns []logstorage.BlockColumn, rowIdx int) string {
//line app/vlselect/logsql/query_response.qtpl:18
	qb422016 := qt422016.AcquireByteBuffer()
//line app/vlselect/logsql/query_response.qtpl:18
	WriteJSONRow(qb422016, columns, rowIdx)
//line app/vlselect/logsql/query_response.qtpl:18
	qs422016 := string(qb422016.B)
//line app/vlselect/logsql/query_response.qtpl:18
	qt422016.ReleaseByteBuffer(qb422016)
//line app/vlselect/logsql/query_response.qtpl:18
	return qs422016
//line app/vlselect/logsql/query_response.qtpl:18
}

// JSONRows prints formatted rows

//line app/vlselect/logsql/query_response.qtpl:21
func StreamJSONRows(qw422016 *qt422016.Writer, rows [][]logstorage.Field) {
//line app/vlselect/logsql/query_response.qtpl:22
	if len(rows) == 0 {
//line app/vlselect/logsql/query_response.qtpl:23
		return
//line app/vlselect/logsql/query_response.qtpl:24
	}
//line app/vlselect/logsql/query_response.qtpl:25
	for _, fields := range rows {
//line app/vlselect/logsql/query_response.qtpl:25
		qw422016.N().S(`{`)
//line app/vlselect/logsql/query_response.qtpl:27
		if len(fields) > 0 {
//line app/vlselect/logsql/query_response.qtpl:29
			f := fields[0]
			fields = fields[1:]

//line app/vlselect/logsql/query_response.qtpl:32
			qw422016.N().Q(f.Name)
//line app/vlselect/logsql/query_response.qtpl:32
			qw422016.N().S(`:`)
//line app/vlselect/logsql/query_response.qtpl:32
			qw422016.N().Q(f.Value)
//line app/vlselect/logsql/query_response.qtpl:33
			for _, f := range fields {
//line app/vlselect/logsql/query_response.qtpl:33
				qw422016.N().S(`,`)
//line app/vlselect/logsql/query_response.qtpl:34
				qw422016.N().Q(f.Name)
//line app/vlselect/logsql/query_response.qtpl:34
				qw422016.N().S(`:`)
//line app/vlselect/logsql/query_response.qtpl:34
				qw422016.N().Q(f.Value)
//line app/vlselect/logsql/query_response.qtpl:35
			}
//line app/vlselect/logsql/query_response.qtpl:36
		}
//line app/vlselect/logsql/query_response.qtpl:36
		qw422016.N().S(`}`)
//line app/vlselect/logsql/query_response.qtpl:37
		qw422016.N().S(`
`)
//line app/vlselect/logsql/query_response.qtpl:38
	}
//line app/vlselect/logsql/query_response.qtpl:39
}

//line app/vlselect/logsql/query_response.qtpl:39
func WriteJSONRows(qq422016 qtio422016.Writer, rows [][]logstorage.Field) {
//line app/vlselect/logsql/query_response.qtpl:39
	qw422016 := qt422016.AcquireWriter(qq422016)
//line app/vlselect/logsql/query_response.qtpl:39
	StreamJSONRows(qw422016, rows)
//line app/vlselect/logsql/query_response.qtpl:39
	qt422016.ReleaseWriter(qw422016)
//line app/vlselect/logsql/query_response.qtpl:39
}

//line app/vlselect/logsql/query_response.qtpl:39
func JSONRows(rows [][]logstorage.Field) string {
//line app/vlselect/logsql/query_response.qtpl:39
	qb422016 := qt422016.AcquireByteBuffer()
//line app/vlselect/logsql/query_response.qtpl:39
	WriteJSONRows(qb422016, rows)
//line app/vlselect/logsql/query_response.qtpl:39
	qs422016 := string(qb422016.B)
//line app/vlselect/logsql/query_response.qtpl:39
	qt422016.ReleaseByteBuffer(qb422016)
//line app/vlselect/logsql/query_response.qtpl:39
	return qs422016
//line app/vlselect/logsql/query_response.qtpl:39
}
