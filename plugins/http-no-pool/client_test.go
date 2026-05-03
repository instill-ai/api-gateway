package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestClientRegisterer_Name(t *testing.T) {
	if got := string(ClientRegisterer); got != "http-no-pool-client" {
		t.Fatalf("expected http-no-pool-client, got %s", got)
	}
}

func TestRegisterClients_WrongName(t *testing.T) {
	var handler http.Handler
	ClientRegisterer.RegisterClients(func(name string, factory func(context.Context, map[string]any) (http.Handler, error)) {
		h, err := factory(context.Background(), map[string]any{"name": "wrong-name"})
		if err == nil {
			t.Fatal("expected error for wrong name")
		}
		handler = h
	})
	if handler != nil {
		t.Fatal("handler should be nil")
	}
}

func TestRegisterClients_PerRequestConnection(t *testing.T) {
	var connCount atomic.Int64
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connCount.Add(1)
		w.Header().Set("X-Test", "ok")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer backend.Close()

	var handler http.Handler
	ClientRegisterer.RegisterClients(func(name string, factory func(context.Context, map[string]any) (http.Handler, error)) {
		h, err := factory(context.Background(), map[string]any{
			"name":                "http-no-pool-client",
			"http-no-pool-client": map[string]any{},
		})
		if err != nil {
			t.Fatal(err)
		}
		handler = h
	})

	if handler == nil {
		t.Fatal("handler is nil")
	}

	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest("GET", backend.URL+"/v1alpha/health", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		if string(body) != `{"status":"ok"}` {
			t.Fatalf("request %d: unexpected body %q", i, body)
		}
		if resp.Header.Get("X-Test") != "ok" {
			t.Fatalf("request %d: X-Test header not forwarded", i)
		}
	}

	if got := connCount.Load(); got < 5 {
		t.Fatalf("expected at least 5 connections (one per request), got %d — connection pooling is leaking", got)
	}
}
