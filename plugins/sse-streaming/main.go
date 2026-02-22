package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

/*

the following configuration must be present in the krakend.json file:

"extra_config": {
    "plugin/http-server": {
		"name": ["sse-streaming"],
		"sse-streaming": {
			"backend_host": "http://localhost:9081"
		}
	}
}

the provided endpoint overwrites any existing endpoint in the configuration file.

*/

// pluginName is the name of the plugin, used as a key in the configuration map.
var pluginName = "sse-streaming"

// HandlerRegisterer is the symbol the plugin loader will try to load. It must implement the Registerer interface.
var HandlerRegisterer = registerer(pluginName)

type registerer string

// RegisterHandlers registers the handler function with the given name.
func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]any, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

// registerHandlers extracts configuration and sets up the HTTP handler.
func (r registerer) registerHandlers(ctx context.Context, extra map[string]any, h http.Handler) (http.Handler, error) {
	config, ok := extra[pluginName].(map[string]any)
	if !ok {
		return h, errors.New("configuration not found")
	}

	// Extract configuration values
	backendHost, backendHostOk := config["backend_host"].(string)

	// Check if all required configuration values are present
	if !backendHostOk {
		return h, errors.New("missing required configuration values")
	}

	if backendHost == "" {
		return h, errors.New("backend_host cannot be empty")
	}

	// Validate the backend host
	if _, err := url.ParseRequestURI(backendHost); err != nil {
		return h, errors.New("invalid backend_host URL")
	}

	// Create HTTP client with OpenTelemetry instrumentation
	httpClient := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	// Return a new HTTP handler that wraps the original handler with custom logic.
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Extract OpenTelemetry context from incoming request
		otelCtx := req.Context()

		// Create a span for the SSE streaming plugin
		tracer := trace.SpanFromContext(otelCtx).TracerProvider().Tracer("sse-streaming")
		spanCtx, span := tracer.Start(otelCtx, "sse-streaming.handle_request",
			trace.WithAttributes(
				attribute.String("http.method", req.Method),
				attribute.String("http.url", req.URL.String()),
				attribute.String("http.user_agent", req.UserAgent()),
				attribute.String("plugin.name", pluginName),
			),
		)
		defer span.End()

		// Add span context to request
		req = req.WithContext(spanCtx)

		// This is a quick solution since we only support sse for pipeline trigger endpoint
		if req.Header.Get("Accept") == "text/event-stream" {
			authType := req.Header.Get("Instill-Auth-Type")
			if authType == "" || authType == "visitor" {
				span.SetAttributes(attribute.String("sse.action", "rejected_unauthenticated"))
				span.SetStatus(codes.Error, "unauthenticated SSE request")
				http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
				return
			}
			span.SetAttributes(attribute.String("sse.action", "proxy_sse_request"))
			proxyHandler(spanCtx, w, req, httpClient, backendHost)
		} else {
			span.SetAttributes(attribute.String("sse.action", "pass_through"))
			h.ServeHTTP(w, req)
		}

	}), nil
}

// proxyHandler forwards the request to the actual SSE server and streams the response back to the client.
func proxyHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, httpClient http.Client, backendHost string) {
	// Create a child span for the proxy operation
	tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer("sse-streaming")
	proxySpanCtx, proxySpan := tracer.Start(ctx, "sse-streaming.proxy_request",
		trace.WithAttributes(
			attribute.String("sse.action", "proxy_to_backend"),
		),
	)
	defer proxySpan.End()

	url := string(r.URL.Path)
	url = strings.ReplaceAll(url, "/internal", "")
	targetURL := fmt.Sprintf("http://%s%s", backendHost, url)

	proxySpan.SetAttributes(
		attribute.String("sse.backend_host", backendHost),
		attribute.String("sse.target_url", targetURL),
		attribute.String("sse.original_path", r.URL.Path),
		attribute.String("sse.modified_path", url),
	)

	req, err := http.NewRequest("POST", targetURL, r.Body)
	if err != nil {
		proxySpan.RecordError(err)
		proxySpan.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header = r.Header

	// Create a child span for the HTTP request to backend
	httpSpanCtx, httpSpan := tracer.Start(proxySpanCtx, "sse-streaming.http_backend_request",
		trace.WithAttributes(
			attribute.String("http.target", targetURL),
			attribute.String("http.method", "POST"),
		),
	)
	defer httpSpan.End()

	resp, err := httpClient.Do(req.WithContext(httpSpanCtx))
	if err != nil {
		httpSpan.RecordError(err)
		httpSpan.SetStatus(codes.Error, err.Error())
		proxySpan.RecordError(err)
		proxySpan.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httpSpan.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
	proxySpan.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create a child span for streaming
	_, streamSpan := tracer.Start(proxySpanCtx, "sse-streaming.stream_response",
		trace.WithAttributes(
			attribute.String("sse.action", "stream_events"),
		),
	)
	defer streamSpan.End()

	// Create a buffered reader to read the SSE stream
	reader := bufio.NewReader(resp.Body)
	eventCount := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			// End of stream or error
			if eventCount > 0 {
				streamSpan.SetAttributes(attribute.Int("sse.events_sent", eventCount))
			}
			break
		}

		// Write the event to the client
		_, err = w.Write([]byte(line))
		if err != nil {
			streamSpan.RecordError(err)
			streamSpan.SetStatus(codes.Error, err.Error())
			break
		}

		eventCount++

		// Flush the response writer to ensure the event is sent immediately
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}

	streamSpan.SetAttributes(attribute.Int("sse.events_sent", eventCount))
	proxySpan.SetAttributes(attribute.Int("sse.total_events", eventCount))
}

func main() {}

// This logger is replaced by the RegisterLogger method to load the one from KrakenD
var logger Logger = noopLogger{}

func (registerer) RegisterLogger(v any) {
	l, ok := v.(Logger)
	if !ok {
		return
	}
	logger = l
	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Logger loaded", HandlerRegisterer))
}

// Logger is the interface for the logger
type Logger interface {
	Debug(v ...any)
	Info(v ...any)
	Warning(v ...any)
	Error(v ...any)
	Critical(v ...any)
	Fatal(v ...any)
}

// Empty logger implementation
type noopLogger struct{}

func (n noopLogger) Debug(_ ...any)    {}
func (n noopLogger) Info(_ ...any)     {}
func (n noopLogger) Warning(_ ...any)  {}
func (n noopLogger) Error(_ ...any)    {}
func (n noopLogger) Critical(_ ...any) {}
func (n noopLogger) Fatal(_ ...any)    {}
