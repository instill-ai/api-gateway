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
		"name": ["pipeline-sse-streaming"],
		"pipeline-sse-streaming": {
			"backend_host": "http://localhost:9081"
		}
	}
}

the provided endpoint overwrites any existing endpoint in the configuration file.

*/

var pluginName = "pipeline-sse-streaming"

var HandlerRegisterer = registerer(pluginName)

type registerer string

func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]any, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

func (r registerer) registerHandlers(ctx context.Context, extra map[string]any, h http.Handler) (http.Handler, error) {
	config, ok := extra[pluginName].(map[string]any)
	if !ok {
		return h, errors.New("configuration not found")
	}

	backendHost, backendHostOk := config["backend_host"].(string)

	if !backendHostOk {
		return h, errors.New("missing required configuration values")
	}

	if backendHost == "" {
		return h, errors.New("backend_host cannot be empty")
	}

	if _, err := url.ParseRequestURI(backendHost); err != nil {
		return h, errors.New("invalid backend_host URL")
	}

	httpClient := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		otelCtx := req.Context()

		tracer := trace.SpanFromContext(otelCtx).TracerProvider().Tracer(pluginName)
		spanCtx, span := tracer.Start(otelCtx, pluginName+".handle_request",
			trace.WithAttributes(
				attribute.String("http.method", req.Method),
				attribute.String("http.url", req.URL.String()),
				attribute.String("http.user_agent", req.UserAgent()),
				attribute.String("plugin.name", pluginName),
			),
		)
		defer span.End()

		req = req.WithContext(spanCtx)

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

func proxyHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, httpClient http.Client, backendHost string) {
	tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer(pluginName)
	proxySpanCtx, proxySpan := tracer.Start(ctx, pluginName+".proxy_request",
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

	httpSpanCtx, httpSpan := tracer.Start(proxySpanCtx, pluginName+".http_backend_request",
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

	_, streamSpan := tracer.Start(proxySpanCtx, pluginName+".stream_response",
		trace.WithAttributes(
			attribute.String("sse.action", "stream_events"),
		),
	)
	defer streamSpan.End()

	reader := bufio.NewReader(resp.Body)
	eventCount := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if eventCount > 0 {
				streamSpan.SetAttributes(attribute.Int("sse.events_sent", eventCount))
			}
			break
		}

		_, err = w.Write([]byte(line))
		if err != nil {
			streamSpan.RecordError(err)
			streamSpan.SetStatus(codes.Error, err.Error())
			break
		}

		eventCount++

		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}

	streamSpan.SetAttributes(attribute.Int("sse.events_sent", eventCount))
	proxySpan.SetAttributes(attribute.Int("sse.total_events", eventCount))
}

func main() {}

var logger Logger = noopLogger{}

func (registerer) RegisterLogger(v any) {
	l, ok := v.(Logger)
	if !ok {
		return
	}
	logger = l
	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Logger loaded", HandlerRegisterer))
}

type Logger interface {
	Debug(v ...any)
	Info(v ...any)
	Warning(v ...any)
	Error(v ...any)
	Critical(v ...any)
	Fatal(v ...any)
}

type noopLogger struct{}

func (n noopLogger) Debug(_ ...any)    {}
func (n noopLogger) Info(_ ...any)     {}
func (n noopLogger) Warning(_ ...any)  {}
func (n noopLogger) Error(_ ...any)    {}
func (n noopLogger) Critical(_ ...any) {}
func (n noopLogger) Fatal(_ ...any)    {}
