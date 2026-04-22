package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// proxyTestCase drives proxyToMinIO against a captured fake MinIO and asserts
// the request that reached the upstream, plus the response that reached the
// client. Table-driven so each BLOB-INV-RANGE permutation is a single row.
type proxyTestCase struct {
	name string

	// clientReq describes the incoming browser request.
	reqMethod  string
	reqHeaders http.Header

	// fake MinIO behaviour.
	upstreamStatus  int
	upstreamHeaders http.Header
	upstreamBody    string

	// mode: base64-presigned (false) vs object-ID (true).
	isObjectID bool

	// assertions against the captured upstream request.
	wantUpstreamHeaders   map[string]string // must be present with exact value
	wantUpstreamAbsent    []string          // must not be sent upstream
	wantClientStatus      int
	wantClientBody        string
	wantClientCacheHeader string // "" means "must NOT be set"
}

func newNoopTracer() trace.Tracer {
	return noop.NewTracerProvider().Tracer("blob-test")
}

func TestProxyToMinIO_HeaderForwardingAndCacheScoping(t *testing.T) {
	cases := []proxyTestCase{
		{
			name:      "range_forwarded_and_206_passes_through",
			reqMethod: http.MethodGet,
			reqHeaders: http.Header{
				"Range":  []string{"bytes=0-1023"},
				"Accept": []string{"video/mp4"},
			},
			upstreamStatus: http.StatusPartialContent,
			upstreamHeaders: http.Header{
				"Content-Range":  []string{"bytes 0-1023/5000"},
				"Content-Length": []string{"1024"},
				"Accept-Ranges":  []string{"bytes"},
				"Content-Type":   []string{"video/mp4"},
			},
			upstreamBody: strings.Repeat("x", 1024),
			isObjectID:   true,
			wantUpstreamHeaders: map[string]string{
				"Range":  "bytes=0-1023",
				"Accept": "video/mp4",
			},
			wantClientStatus: http.StatusPartialContent,
			wantClientBody:   strings.Repeat("x", 1024),
			// BLOB-INV-RANGE: 206 must never carry the aggressive 24h
			// Cache-Control; otherwise the browser cache would treat a
			// partial body as the whole object.
			wantClientCacheHeader: "",
		},
		{
			name:      "if_range_forwarded",
			reqMethod: http.MethodGet,
			reqHeaders: http.Header{
				"Range":    []string{"bytes=1000-1999"},
				"If-Range": []string{`"etag-abc"`},
			},
			upstreamStatus: http.StatusPartialContent,
			upstreamHeaders: http.Header{
				"Content-Range": []string{"bytes 1000-1999/5000"},
			},
			upstreamBody: strings.Repeat("y", 1000),
			isObjectID:   true,
			wantUpstreamHeaders: map[string]string{
				"Range":    "bytes=1000-1999",
				"If-Range": `"etag-abc"`,
			},
			wantClientStatus:      http.StatusPartialContent,
			wantClientBody:        strings.Repeat("y", 1000),
			wantClientCacheHeader: "",
		},
		{
			name:      "if_none_match_forwarded_and_304_passes_through",
			reqMethod: http.MethodGet,
			reqHeaders: http.Header{
				"If-None-Match": []string{`"etag-abc"`},
			},
			upstreamStatus: http.StatusNotModified,
			upstreamHeaders: http.Header{
				"ETag": []string{`"etag-abc"`},
			},
			upstreamBody: "",
			isObjectID:   true,
			wantUpstreamHeaders: map[string]string{
				"If-None-Match": `"etag-abc"`,
			},
			wantClientStatus: http.StatusNotModified,
			wantClientBody:   "",
			// 304 already carries its own cache semantics via the original
			// 200's Cache-Control; don't override.
			wantClientCacheHeader: "",
		},
		{
			name:      "if_modified_since_forwarded",
			reqMethod: http.MethodGet,
			reqHeaders: http.Header{
				"If-Modified-Since": []string{"Wed, 21 Oct 2015 07:28:00 GMT"},
			},
			upstreamStatus: http.StatusNotModified,
			isObjectID:     true,
			wantUpstreamHeaders: map[string]string{
				"If-Modified-Since": "Wed, 21 Oct 2015 07:28:00 GMT",
			},
			wantClientStatus:      http.StatusNotModified,
			wantClientCacheHeader: "",
		},
		{
			name:       "no_range_200_object_id_mode_adds_cache_control",
			reqMethod:  http.MethodGet,
			reqHeaders: http.Header{"Accept": []string{"video/mp4"}},
			upstreamStatus: http.StatusOK,
			upstreamHeaders: http.Header{
				"Content-Type":   []string{"video/mp4"},
				"Content-Length": []string{"11"},
			},
			upstreamBody: "full-object",
			isObjectID:   true,
			wantUpstreamHeaders: map[string]string{
				"Accept": "video/mp4",
			},
			// BLOB-INV-RANGE: when no Range was supplied, nothing is sent
			// upstream either.
			wantUpstreamAbsent:    []string{"Range", "If-Range", "If-None-Match", "If-Modified-Since"},
			wantClientStatus:      http.StatusOK,
			wantClientBody:        "full-object",
			wantClientCacheHeader: "public, max-age=86400",
		},
		{
			name:                  "legacy_presigned_url_mode_never_adds_cache_control",
			reqMethod:             http.MethodGet,
			reqHeaders:            http.Header{},
			upstreamStatus:        http.StatusOK,
			upstreamBody:          "legacy",
			isObjectID:            false,
			wantClientStatus:      http.StatusOK,
			wantClientBody:        "legacy",
			wantClientCacheHeader: "",
		},
		{
			name:      "legacy_presigned_url_with_range_still_forwards",
			reqMethod: http.MethodGet,
			reqHeaders: http.Header{
				"Range": []string{"bytes=0-99"},
			},
			upstreamStatus: http.StatusPartialContent,
			upstreamHeaders: http.Header{
				"Content-Range": []string{"bytes 0-99/500"},
			},
			upstreamBody: strings.Repeat("z", 100),
			isObjectID:   false,
			wantUpstreamHeaders: map[string]string{
				"Range": "bytes=0-99",
			},
			wantClientStatus:      http.StatusPartialContent,
			wantClientBody:        strings.Repeat("z", 100),
			wantClientCacheHeader: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var captured *http.Request
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				captured = r.Clone(context.Background())
				for k, vs := range tc.upstreamHeaders {
					for _, v := range vs {
						w.Header().Add(k, v)
					}
				}
				status := tc.upstreamStatus
				if status == 0 {
					status = http.StatusOK
				}
				w.WriteHeader(status)
				if tc.upstreamBody != "" {
					_, _ = io.WriteString(w, tc.upstreamBody)
				}
			}))
			defer backend.Close()

			blobURL, err := url.Parse(backend.URL)
			if err != nil {
				t.Fatalf("parsing backend URL: %v", err)
			}

			clientReq := httptest.NewRequest(tc.reqMethod, "/v1alpha/blob-urls/stub", nil)
			for k, vs := range tc.reqHeaders {
				for _, v := range vs {
					clientReq.Header.Add(k, v)
				}
			}

			rec := httptest.NewRecorder()
			proxyToMinIO(
				context.Background(),
				newNoopTracer(),
				&http.Client{},
				rec,
				clientReq,
				blobURL,
				tc.isObjectID,
			)

			if captured == nil {
				t.Fatalf("fake MinIO never received a request")
			}

			for h, want := range tc.wantUpstreamHeaders {
				if got := captured.Header.Get(h); got != want {
					t.Errorf("upstream header %q: got %q, want %q", h, got, want)
				}
			}
			for _, h := range tc.wantUpstreamAbsent {
				if got := captured.Header.Get(h); got != "" {
					t.Errorf("upstream header %q must be absent, got %q", h, got)
				}
			}

			if got := rec.Code; got != tc.wantClientStatus {
				t.Errorf("client status: got %d, want %d", got, tc.wantClientStatus)
			}
			if got := rec.Body.String(); got != tc.wantClientBody {
				t.Errorf("client body length: got %d, want %d", len(got), len(tc.wantClientBody))
			}

			gotCache := rec.Header().Get("Cache-Control")
			if tc.wantClientCacheHeader == "" {
				if gotCache != "" {
					t.Errorf("Cache-Control must not be set on this response, got %q", gotCache)
				}
			} else if gotCache != tc.wantClientCacheHeader {
				t.Errorf("Cache-Control: got %q, want %q", gotCache, tc.wantClientCacheHeader)
			}
		})
	}
}

// TestProxyToMinIO_PropagatesUpstreamResponseHeaders asserts that headers the
// browser needs to honour the 206 / 304 protocol (Content-Range, Accept-Ranges,
// ETag, Last-Modified) flow back unchanged from MinIO to the client.
func TestProxyToMinIO_PropagatesUpstreamResponseHeaders(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Range", "bytes 0-9/100")
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("ETag", `"etag-xyz"`)
		w.Header().Set("Last-Modified", "Wed, 21 Oct 2015 07:28:00 GMT")
		// Access-Control-Allow-Origin is intentionally stripped by the proxy —
		// CORS is set by KrakenD's CORS plugin.
		w.Header().Set("Access-Control-Allow-Origin", "https://somewhere-else")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = io.WriteString(w, "0123456789")
	}))
	defer backend.Close()

	blobURL, _ := url.Parse(backend.URL)
	req := httptest.NewRequest(http.MethodGet, "/v1alpha/blob-urls/stub", nil)
	req.Header.Set("Range", "bytes=0-9")

	rec := httptest.NewRecorder()
	proxyToMinIO(context.Background(), newNoopTracer(), &http.Client{}, rec, req, blobURL, true)

	if got := rec.Code; got != http.StatusPartialContent {
		t.Fatalf("status: got %d, want 206", got)
	}
	for _, pair := range [][2]string{
		{"Content-Range", "bytes 0-9/100"},
		{"Accept-Ranges", "bytes"},
		{"Etag", `"etag-xyz"`},
		{"Last-Modified", "Wed, 21 Oct 2015 07:28:00 GMT"},
	} {
		if got := rec.Header().Get(pair[0]); got != pair[1] {
			t.Errorf("response header %q: got %q, want %q", pair[0], got, pair[1])
		}
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin must be stripped, got %q", got)
	}
	if got := rec.Header().Get("Cache-Control"); got != "" {
		t.Errorf("Cache-Control must not be set on 206 responses, got %q", got)
	}
}

// TestProxyToMinIO_HEADRequestForwardsMethod guards that the proxy preserves
// the HTTP method, which HLS/video players rely on for metadata probes.
func TestProxyToMinIO_HEADRequestForwardsMethod(t *testing.T) {
	var capturedMethod string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		w.Header().Set("Content-Length", "12345")
		w.Header().Set("Accept-Ranges", "bytes")
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	blobURL, _ := url.Parse(backend.URL)
	req := httptest.NewRequest(http.MethodHead, "/v1alpha/blob-urls/stub", nil)

	rec := httptest.NewRecorder()
	proxyToMinIO(context.Background(), newNoopTracer(), &http.Client{}, rec, req, blobURL, true)

	if capturedMethod != http.MethodHead {
		t.Errorf("upstream method: got %q, want HEAD", capturedMethod)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Accept-Ranges"); got != "bytes" {
		t.Errorf("Accept-Ranges: got %q, want bytes", got)
	}
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=86400" {
		t.Errorf("Cache-Control: got %q, want public, max-age=86400", got)
	}
}

func TestTryDecodeAsPresignedURL(t *testing.T) {
	cases := []struct {
		name    string
		encoded string
		wantNil bool
	}{
		{
			name:    "valid_base64_url",
			encoded: "aHR0cDovL21pbmlvOjkwMDAvYnVja2V0L29iamVjdA==",
			wantNil: false,
		},
		{
			name:    "plain_object_id_is_not_decoded",
			encoded: "obj-abc123",
			wantNil: true,
		},
		{
			name:    "empty_string",
			encoded: "",
			wantNil: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tryDecodeAsPresignedURL(tc.encoded)
			if (got == nil) != tc.wantNil {
				t.Errorf("tryDecodeAsPresignedURL(%q): got %v, wantNil=%v", tc.encoded, got, tc.wantNil)
			}
		})
	}
}
