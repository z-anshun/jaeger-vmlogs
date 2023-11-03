package store

import (
	"encoding/json"
	"github.com/jaegertracing/jaeger/model"
	"testing"
)

func TestName(t *testing.T) {
	span := &model.Span{}

	json.Unmarshal([]byte(str), span)
	span.Process = &model.Process{
		ServiceName: "123",
		Tags:        nil,
	}
	p := GetParser()
	p.ParseToTraceMsg(span)
	t.Log(p.Fields)
	PutParser(p)
}

var str = `{
                    "traceID": "4c6128967d3943e1abc6a52c45538ce3",
                    "spanID": "44153e7f1ba7628d",
                    "operationName": "TZone.ad-assigner/AssignerService_GetAds",
                    "references": [
                        {
                            "refType": "CHILD_OF",
                            "traceID": "4c6128967d3943e1abc6a52c45538ce3",
                            "spanID": "6aada9037acf7de6"
                        }
                    ],
                    "startTime": 1698739239415411,
                    "duration": 15171,
                    "tags": [
                        {
                            "key": "component",
                            "type": "string",
                            "value": "search-words-web"
                        },
                        {
                            "key": "rpc.method",
                            "type": "string",
                            "value": "AssignerService_GetAds"
                        },
                        {
                            "key": "rpc.system",
                            "type": "string",
                            "value": "TZone"
                        },
                        {
                            "key": "rpc.service",
                            "type": "string",
                            "value": "ad-assigner"
                        },
                        {
                            "key": "trace.tail_sampling.policy_evaluated",
                            "type": "bool",
                            "value": true
                        }
                    ],
                    "logs": [],
                    "processID": "p1",
                    "warnings": null
                }`
