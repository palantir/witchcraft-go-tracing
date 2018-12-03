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

package b3_test

import (
	"net/http"
	"testing"

	"github.com/palantir/witchcraft-go-tracing/wtracing"
	"github.com/palantir/witchcraft-go-tracing/wtracing/propagation/b3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpanInjector(t *testing.T) {
	for _, tc := range []struct {
		name           string
		sc             wtracing.SpanContext
		wantHeaderVals map[string]string
	}{
		{
			name: "full span context injection",
			sc: wtracing.SpanContext{
				TraceID:  idHexVal,
				ID:       idHexVal,
				ParentID: (*wtracing.SpanID)(strPtr(otherIDHexVal)),
				Debug:    false,
				Sampled:  boolPtr(true),
			},
			wantHeaderVals: map[string]string{
				"X-B3-TraceId":      idHexVal,
				"X-B3-SpanId":       idHexVal,
				"X-B3-ParentSpanId": otherIDHexVal,
				"X-B3-Sampled":      "1",
			},
		},
		{
			name: "ids only injected for valid span, but non-id values still injected",
			sc: wtracing.SpanContext{
				TraceID: idHexVal,
				// SpanID is missing, so span is not valid
				ParentID: (*wtracing.SpanID)(strPtr(otherIDHexVal)),
				Debug:    false,
				Sampled:  boolPtr(true),
			},
			wantHeaderVals: map[string]string{
				"X-B3-Sampled": "1",
			},
		},
		{
			name: "injecting span with debug true sets sampled to off",
			sc: wtracing.SpanContext{
				TraceID:  idHexVal,
				ID:       idHexVal,
				ParentID: (*wtracing.SpanID)(strPtr(otherIDHexVal)),
				Debug:    true,
				Sampled:  boolPtr(true),
			},
			wantHeaderVals: map[string]string{
				"X-B3-TraceId":      idHexVal,
				"X-B3-SpanId":       idHexVal,
				"X-B3-ParentSpanId": otherIDHexVal,
				"X-B3-Flags":        "1",
			},
		},
		{
			name: "sampled header not explicitly set if not defined in original span",
			sc: wtracing.SpanContext{
				TraceID: idHexVal,
				ID:      idHexVal,
			},
			wantHeaderVals: map[string]string{
				"X-B3-TraceId": idHexVal,
				"X-B3-SpanId":  idHexVal,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "", nil)
			require.NoError(t, err)
			b3.SpanInjector(req)(tc.sc)
			gotHeader := req.Header

			wantHeader := http.Header{}
			for k, v := range tc.wantHeaderVals {
				wantHeader.Set(k, v)
			}
			assert.Equal(t, wantHeader, gotHeader)
		})
	}
}
