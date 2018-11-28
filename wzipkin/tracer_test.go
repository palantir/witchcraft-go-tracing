// Copyright (c) 2018 Palantir Technologies. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wzipkin_test

import (
	"testing"

	"github.com/palantir/witchcraft-go-tracing/wtracing"
	"github.com/palantir/witchcraft-go-tracing/wzipkin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTracerStartSpan(t *testing.T) {
	reporterMap := make(map[string]interface{})

	tracer, err := wzipkin.NewTracer(&testReporter{
		reporterMap: reporterMap,
	})
	require.NoError(t, err)

	span := tracer.StartSpan("mySpan")
	span.Finish()

	assert.NotEqual(t, reporterMap["traceID"], wtracing.TraceID(""))
	assert.NotEqual(t, reporterMap["spanID"], wtracing.SpanID(""))
	assert.Equal(t, reporterMap["traceID"], wtracing.TraceID(reporterMap["spanID"].(wtracing.SpanID)))
	assert.Equal(t, reporterMap["parentID"], (*wtracing.SpanID)(nil))
	assert.Equal(t, reporterMap["debug"], false)
	assert.Equal(t, *reporterMap["sampled"].(*bool), true)
	assert.Equal(t, reporterMap["err"], (error)(nil))

	assert.Equal(t, reporterMap["name"], "mySpan")
	assert.Equal(t, reporterMap["kind"], wtracing.Kind(""))
	assert.NotEqual(t, reporterMap["timestamp"], nil)
	assert.NotEqual(t, reporterMap["duration"], nil)
	assert.Equal(t, reporterMap["localEndpoint"].(*wtracing.Endpoint), (*wtracing.Endpoint)(nil))
	assert.Equal(t, reporterMap["remoteEndpoint"].(*wtracing.Endpoint), (*wtracing.Endpoint)(nil))
}

func TestTracerStartChildSpan(t *testing.T) {
	reporterMap := make(map[string]interface{})

	tracer, err := wzipkin.NewTracer(&testReporter{
		reporterMap: reporterMap,
	})
	require.NoError(t, err)

	rootSpan := tracer.StartSpan("myRootSpan")

	childSpan := tracer.StartSpan("myChildSpan", wtracing.WithParent(rootSpan.Context()))
	childSpan.Finish()

	assert.NotEqual(t, reporterMap["traceID"], wtracing.TraceID(""))
	assert.NotEqual(t, reporterMap["spanID"], wtracing.SpanID(""))
	assert.NotEqual(t, reporterMap["traceID"], wtracing.TraceID(reporterMap["spanID"].(wtracing.SpanID)))
	assert.Equal(t, *(reporterMap["parentID"].(*wtracing.SpanID)), (wtracing.SpanID)(reporterMap["traceID"].(wtracing.TraceID)))
	assert.Equal(t, reporterMap["debug"], false)
	assert.Equal(t, *reporterMap["sampled"].(*bool), true)
	assert.Equal(t, reporterMap["err"], (error)(nil))

	assert.Equal(t, reporterMap["name"], "myChildSpan")
	assert.Equal(t, reporterMap["kind"], wtracing.Kind(""))
	assert.NotEqual(t, reporterMap["timestamp"], nil)
	assert.NotEqual(t, reporterMap["duration"], nil)
	assert.Equal(t, reporterMap["localEndpoint"].(*wtracing.Endpoint), (*wtracing.Endpoint)(nil))
	assert.Equal(t, reporterMap["remoteEndpoint"].(*wtracing.Endpoint), (*wtracing.Endpoint)(nil))

	rootSpan.Finish()
}

type testReporter struct {
	reporterMap map[string]interface{}
}

func (r *testReporter) Send(span wtracing.SpanModel) {
	r.reporterMap["traceID"] = span.TraceID
	r.reporterMap["spanID"] = span.ID
	r.reporterMap["parentID"] = span.ParentID
	r.reporterMap["debug"] = span.Debug
	r.reporterMap["sampled"] = span.Sampled
	r.reporterMap["err"] = span.Err

	r.reporterMap["name"] = span.Name
	r.reporterMap["kind"] = span.Kind
	r.reporterMap["timestamp"] = span.Timestamp
	r.reporterMap["duration"] = span.Duration
	r.reporterMap["localEndpoint"] = span.LocalEndpoint
	r.reporterMap["remoteEndpoint"] = span.RemoteEndpoint
}

func (r *testReporter) Close() error {
	return nil
}
