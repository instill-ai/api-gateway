package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/luraproject/lura/v2/logging"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// pluginName is the plugin name
var pluginName = "grpc-proxy-server"

// HandlerRegisterer is the symbol the plugin loader will try to load. It must implement the Registerer interface
var HandlerRegisterer = registerer(pluginName)

type registerer string

type responseHijacker struct {
	w                    http.ResponseWriter
	grpcStatus           string
	grpcMessage          string
	grpcStatusDetailsBin string
}

type Trailer interface {
	WriteTrailer()
}

func NewResponseHijacker(w http.ResponseWriter) http.ResponseWriter {
	return &responseHijacker{w: w}
}

func (l *responseHijacker) Header() http.Header {
	return l.w.Header()
}

func (l *responseHijacker) Write(b []byte) (int, error) {
	return l.w.Write(b)
}

func (l *responseHijacker) WriteHeader(s int) {
	// Note: hijack the trailers and remove them from headers
	l.grpcStatus = l.w.Header().Get("Grpc-Status")
	l.grpcMessage = l.w.Header().Get("Grpc-Message")
	l.grpcStatusDetailsBin = l.w.Header().Get("Grpc-Status-Details-Bin")
	l.w.Header().Del("Grpc-Status")
	l.w.Header().Del("Grpc-Message")
	l.w.Header().Del("Grpc-Status-Details-Bin")

	l.w.Header().Set("Trailer", "Grpc-Status")
	if l.grpcMessage != "" {
		l.w.Header().Add("Trailer", "Grpc-Message")
	}
	if l.grpcStatusDetailsBin != "" {
		l.w.Header().Add("Trailer", "Grpc-Status-Details-Bin")
	}

	l.w.WriteHeader(s)
}

func (l *responseHijacker) WriteTrailer() {
	l.w.Header().Set("Grpc-Status", l.grpcStatus)
	if l.grpcMessage != "" {
		l.w.Header().Set("Grpc-Message", l.grpcMessage)
	}
	if l.grpcStatusDetailsBin != "" {
		l.w.Header().Set("Grpc-Status-Details-Bin", l.grpcStatusDetailsBin)
	}
}

func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]any, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

func (r registerer) registerHandlers(ctx context.Context, extra map[string]any, h http.Handler) (http.Handler, error) {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Extract OpenTelemetry context from incoming request
		otelCtx := req.Context()

		// Create a span for the gRPC proxy plugin
		tracer := trace.SpanFromContext(otelCtx).TracerProvider().Tracer("grpc-proxy-server")
		spanCtx, span := tracer.Start(otelCtx, "grpc-proxy.handle_request",
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

		if req.Header.Get("Accept") == "text/event-stream" {
			// For SSE, we need to skip this plugin.
			span.SetAttributes(attribute.String("grpc.proxy.skip_reason", "sse_request"))
			h.ServeHTTP(w, req)
		} else {
			// Create a child span for gRPC processing
			grpcSpanCtx, grpcSpan := tracer.Start(spanCtx, "grpc-proxy.process_grpc",
				trace.WithAttributes(
					attribute.String("grpc.proxy.action", "process_grpc_request"),
				),
			)
			defer grpcSpan.End()

			w = NewResponseHijacker(w)
			// Pass the OpenTelemetry context to the next handler
			req = req.WithContext(grpcSpanCtx)

			// Add response hijacker info to span
			grpcSpan.SetAttributes(attribute.String("grpc.proxy.response_hijacker", "enabled"))

			h.ServeHTTP(w, req)

			// Record gRPC status in span
			if hijacker, ok := w.(*responseHijacker); ok {
				grpcSpan.SetAttributes(
					attribute.String("grpc.status", hijacker.grpcStatus),
					attribute.String("grpc.message", hijacker.grpcMessage),
				)
			}

			w.(Trailer).WriteTrailer()
		}
	}), &http2.Server{}), nil
}

func (registerer) RegisterLogger(v any) {
	l, ok := v.(logging.BasicLogger)
	if !ok {
		return
	}
	logger = l
	logger.Info(fmt.Sprintf("[PLUGIN: %s] Logger loaded", HandlerRegisterer))
}
