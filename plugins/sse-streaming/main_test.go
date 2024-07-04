package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock HTTP server to simulate the backend SSE server
func mockSSEServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data: event1\n\n"))
		w.Write([]byte("data: event2\n\n"))
	}))
}

func TestProxyHandler(t *testing.T) {
	tests := []struct {
		serverURL string
		wantCode  int
		wantBody  string
	}{
		// Positive case: valid SSE server response
		{
			serverURL: "/valid",
			wantCode:  http.StatusOK,
			wantBody:  "data: event1\n\ndata: event2\n\n",
		},
		// Negative case: invalid SSE server URL
		{
			serverURL: "/invalid",
			wantCode:  http.StatusInternalServerError,
			wantBody:  "Failed to connect to downstream SSE server\n",
		},
	}

	// Create a mock SSE server
	mockServer := mockSSEServer()
	defer mockServer.Close()

	for _, tt := range tests {
		t.Run(tt.serverURL, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com"+tt.serverURL, nil)
			w := httptest.NewRecorder()

			// Adjust the serverURL for the positive case to point to the mock server
			if tt.serverURL == "/valid" {
				tt.serverURL = mockServer.URL
			}

			proxyHandler(w, req, tt.serverURL)

			resp := w.Result()
			body, _ := io.ReadAll(resp.Body)

			if resp.StatusCode != tt.wantCode {
				t.Errorf("status code = %v; want %v", resp.StatusCode, tt.wantCode)
			}
			if string(body) != tt.wantBody {
				t.Errorf("body = %v; want %v", string(body), tt.wantBody)
			}
		})
	}
}

func TestMatchStrings(t *testing.T) {
	tests := []struct {
		pattern string
		str     string
		want    bool
		wantID  string
	}{
		// Positive cases
		{pattern: "/api/user/{id}", str: "/api/user/123", want: true, wantID: "123"},
		{pattern: "/product/{id}/details", str: "/product/456/details", want: true, wantID: "456"},

		// Negative cases
		{pattern: "/api/user/{id}", str: "/api/admin/123", want: false, wantID: ""},
		{pattern: "/product/{id}/details", str: "/product/456/info", want: false, wantID: ""},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.str, func(t *testing.T) {
			got, gotID := matchStrings(tt.pattern, tt.str)
			if got != tt.want || gotID != tt.wantID {
				t.Errorf("matchStrings(%q, %q) = %v, %v; want %v, %v", tt.pattern, tt.str, got, gotID, tt.want, tt.wantID)
			}
		})
	}
}

func BenchmarkMatchStrings(b *testing.B) {
	testCases := []struct {
		pattern string
		str     string
	}{
		{"/users/{id}", "/users/123"},
		{"/products/{id}/details", "/products/456/details"},
		{"/orders/{id}/items", "/orders/789/items"},
		{"/categories/{id}/subcategories", "/categories/101/subcategories"},
		{"/users/{id}/posts/{postId}", "/users/202/posts/303"},
	}

	for _, tc := range testCases {
		b.Run(tc.pattern+"_"+tc.str, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				matchStrings(tc.pattern, tc.str)
			}
		})
	}
}
