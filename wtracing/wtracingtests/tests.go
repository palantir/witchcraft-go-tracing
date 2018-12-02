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

package wtracingtests

import (
	"fmt"
	"testing"

	"github.com/palantir/witchcraft-go-tracing/wtracing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ImplProvider struct {
	Name          string
	TracerCreator func(reporter wtracing.Reporter, opts ...wtracing.TracerOption) (wtracing.Tracer, error)
}

type noopFinishSpan wtracing.SpanContext

func (s noopFinishSpan) Context() wtracing.SpanContext {
	return wtracing.SpanContext(s)
}

func (s noopFinishSpan) Finish() {}

func RunTests(t *testing.T, provider ImplProvider) {
	tracer, err := provider.TracerCreator(wtracing.NewNoopReporter())
	require.NoError(t, err)

	t.Run(fmt.Sprintf("%s WithParent", provider.Name), func(t *testing.T) {
		testWithParent(t, tracer)
	})

	t.Run(fmt.Sprintf("%s WithParentSpanContext", provider.Name), func(t *testing.T) {
		testWithParentSpanContext(t, tracer)
	})
}

func testWithParent(t *testing.T, tracer wtracing.Tracer) {
	const idHexVal = "6c2f558d62a7085f"

	t.Run("set valid parent", func(t *testing.T) {
		isSampled := true
		testParentSpan := noopFinishSpan(wtracing.SpanContext{
			TraceID: idHexVal,
			ID:      idHexVal,
			Sampled: &isSampled,
		})

		newSpan := tracer.StartSpan("testSpan", wtracing.WithParent(testParentSpan))

		assert.Equal(t, idHexVal, string(newSpan.Context().TraceID))   // TraceID should be equal to parent
		assert.NotEqual(t, idHexVal, string(newSpan.Context().ID))     // SpanID should be distinct
		assert.Equal(t, idHexVal, string(*newSpan.Context().ParentID)) // ParentID should match parent
	})

	t.Run("set nil parent", func(t *testing.T) {
		isSampled := true
		testParentSpan := noopFinishSpan(wtracing.SpanContext{
			TraceID: idHexVal,
			ID:      idHexVal,
			Sampled: &isSampled,
		})
		newSpan := tracer.StartSpan("testSpan",
			wtracing.WithParent(testParentSpan),
			wtracing.WithParent(nil), // explicitly set nil parent after setting valid parent
		)

		assert.NotEqual(t, idHexVal, string(newSpan.Context().TraceID)) // TraceID should be distinct
		assert.NotEqual(t, idHexVal, string(newSpan.Context().ID))      // SpanID should be distinct
		assert.Nil(t, newSpan.Context().ParentID)                       // ParentID should be nil
	})

	t.Run("set parent with only TraceID", func(t *testing.T) {
		isSampled := true
		testParentSpan := noopFinishSpan(wtracing.SpanContext{
			TraceID: idHexVal,
			Sampled: &isSampled,
		})
		newSpan := tracer.StartSpan("testSpan",
			wtracing.WithParent(testParentSpan),
		)

		assert.Equal(t, idHexVal, string(newSpan.Context().TraceID)) // TraceID should be equal to parent
		assert.Equal(t, idHexVal, string(newSpan.Context().ID))      // SpanID should also be equal to parent (because parent was not valid, this creates a new root span)
		assert.Nil(t, newSpan.Context().ParentID)                    // ParentID should be nil
	})
}

func testWithParentSpanContext(t *testing.T, tracer wtracing.Tracer) {
	const idHexVal = "6c2f558d62a7085f"

	t.Run("set valid parent context", func(t *testing.T) {
		isSampled := true
		testParentSpanCtx := wtracing.SpanContext{
			TraceID: idHexVal,
			ID:      idHexVal,
			Sampled: &isSampled,
		}

		newSpan := tracer.StartSpan("testSpan", wtracing.WithParentSpanContext(testParentSpanCtx))

		assert.Equal(t, idHexVal, string(newSpan.Context().TraceID))   // TraceID should be equal to parent
		assert.NotEqual(t, idHexVal, string(newSpan.Context().ID))     // SpanID should be distinct
		assert.Equal(t, idHexVal, string(*newSpan.Context().ParentID)) // ParentID should match parent
	})

	t.Run("set empty parent context", func(t *testing.T) {
		newSpan := tracer.StartSpan("testSpan",
			wtracing.WithParentSpanContext(wtracing.SpanContext{}),
		)

		assert.NotEqual(t, idHexVal, string(newSpan.Context().TraceID)) // TraceID should be distinct
		assert.NotEqual(t, idHexVal, string(newSpan.Context().ID))      // SpanID should be distinct
		assert.Nil(t, newSpan.Context().ParentID)                       // ParentID should be nil
	})

	t.Run("set parent context with only TraceID", func(t *testing.T) {
		isSampled := true
		testParentSpanContext := wtracing.SpanContext{
			TraceID: idHexVal,
			Sampled: &isSampled,
		}
		newSpan := tracer.StartSpan("testSpan",
			wtracing.WithParentSpanContext(testParentSpanContext),
		)

		assert.Equal(t, idHexVal, string(newSpan.Context().TraceID)) // TraceID should be equal to parent
		assert.Equal(t, idHexVal, string(newSpan.Context().ID))      // SpanID should also be equal to parent (because parent was not valid, this creates a new root span)
		assert.Nil(t, newSpan.Context().ParentID)                    // ParentID should be nil
	})
}
