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

	"github.com/palantir/witchcraft-go-error"
	"github.com/palantir/witchcraft-go-tracing/wtracing"
	"github.com/palantir/witchcraft-go-tracing/wtracing/propagation/b3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	idHexVal      = "6c2f558d62a7085f"
	otherIDHexVal = "7a3e447c51b1244b"
)

func TestSpanExtractor(t *testing.T) {
	for i, tc := range []struct {
		name       string
		headerVals map[string]string
		want       wtracing.SpanContext
	}{
		{
			name: "Values extracted",
			headerVals: map[string]string{
				"X-B3-TraceId":      idHexVal,
				"X-B3-SpanId":       idHexVal,
				"X-B3-ParentSpanId": otherIDHexVal,
				"X-B3-Sampled":      "1",
			},
			want: wtracing.SpanContext{
				TraceID:  idHexVal,
				ID:       idHexVal,
				ParentID: (*wtracing.SpanID)(strPtr(otherIDHexVal)),
				Sampled:  boolPtr(true),
			},
		},
		{
			name: "Error if both TraceID and SpanID absent",
			headerVals: map[string]string{
				"X-B3-Sampled": "1",
			},
			want: wtracing.SpanContext{
				Sampled: boolPtr(true),
				Err:     werror.Error("TraceID missing; SpanID missing"),
			},
		},
		{
			name: "Error if TraceID present but SpanID absent",
			headerVals: map[string]string{
				"X-B3-TraceId": idHexVal,
				"X-B3-Sampled": "1",
			},
			want: wtracing.SpanContext{
				TraceID: idHexVal,
				Sampled: boolPtr(true),
				Err:     werror.Error("SpanID missing"),
			},
		},
		{
			name: "Error if SpanID present but TraceID absent",
			headerVals: map[string]string{
				"X-B3-SpanId":  idHexVal,
				"X-B3-Sampled": "1",
			},
			want: wtracing.SpanContext{
				ID:      idHexVal,
				Sampled: boolPtr(true),
				Err:     werror.Error("TraceID missing"),
			},
		},
		{
			name: "Error if ParentID present when TraceID absent",
			headerVals: map[string]string{
				"X-B3-SpanId":       idHexVal,
				"X-B3-ParentSpanId": otherIDHexVal,
				"X-B3-Sampled":      "1",
			},
			want: wtracing.SpanContext{
				ID:       idHexVal,
				ParentID: (*wtracing.SpanID)(strPtr(otherIDHexVal)),
				Sampled:  boolPtr(true),
				Err:      werror.Error("TraceID missing; ParentID present but TraceID missing"),
			},
		},
		{
			name: "Error if ParentID present when SpanID absent",
			headerVals: map[string]string{
				"X-B3-TraceId":      idHexVal,
				"X-B3-ParentSpanId": otherIDHexVal,
				"X-B3-Sampled":      "1",
			},
			want: wtracing.SpanContext{
				TraceID:  idHexVal,
				ParentID: (*wtracing.SpanID)(strPtr(otherIDHexVal)),
				Sampled:  boolPtr(true),
				Err:      werror.Error("SpanID missing; ParentID present but SpanID missing"),
			},
		},
		{
			name: "Error if ParentID present when TraceID and SpanID absent",
			headerVals: map[string]string{
				"X-B3-ParentSpanId": otherIDHexVal,
				"X-B3-Sampled":      "1",
			},
			want: wtracing.SpanContext{
				ParentID: (*wtracing.SpanID)(strPtr(otherIDHexVal)),
				Sampled:  boolPtr(true),
				Err:      werror.Error("TraceID missing; SpanID missing; ParentID present but TraceID and SpanID missing"),
			},
		},
		{
			name: "Error if sampled value invalid",
			headerVals: map[string]string{
				"X-B3-TraceId": idHexVal,
				"X-B3-SpanId":  idHexVal,
				"X-B3-Sampled": "invalid",
			},
			want: wtracing.SpanContext{
				TraceID: idHexVal,
				ID:      idHexVal,
				Err:     werror.Error("Sampled invalid", werror.SafeParam("sampledHeaderVal", "invalid")),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "localhost", nil)
			require.NoError(t, err)
			for k, v := range tc.headerVals {
				req.Header.Set(k, v)
			}
			got := b3.SpanExtractor(req)()

			// store Err field and set original values to nil so that comparison occurs without the error
			wantErr := tc.want.Err
			tc.want.Err = nil
			gotErr := got.Err
			got.Err = nil

			// verify structs are equal
			assert.Equal(t, tc.want, got, "Case %d", i)
			// verify errors are equal
			werrorsEqual(t, wantErr, gotErr)
		})
	}
}

func werrorsEqual(t *testing.T, wantErr, gotErr error) {
	if wantErr == nil && gotErr == nil {
		return
	} else if wantErr == nil || gotErr == nil {
		assert.Equal(t, wantErr, gotErr)
		return
	}

	assert.Equal(t, wantErr.Error(), gotErr.Error(), "Error messages not equal")

	safeParams1, unsafeParams1 := werror.ParamsFromError(wantErr)
	safeParams2, unsafeParams2 := werror.ParamsFromError(gotErr)

	assert.Equal(t, safeParams1, safeParams2, "SafeParams not equal")
	assert.Equal(t, unsafeParams1, unsafeParams2, "UnsafeParams not equal")
}

func strPtr(in string) *string {
	return &in
}

func boolPtr(in bool) *bool {
	return &in
}
