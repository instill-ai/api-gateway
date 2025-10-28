package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/luraproject/lura/v2/logging"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var pluginName = "blob"

type blob string

func (r blob) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]any, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

func (r blob) registerHandlers(ctx context.Context, extra map[string]any, h http.Handler) (http.Handler, error) {
	config, ok := extra[pluginName].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("configuration not found")
	}

	blobHandler, err := newBlobHandler(config)
	if err != nil {
		return nil, fmt.Errorf("failed to configure handler: %w", err)
	}

	// Create HTTP client with OpenTelemetry instrumentation
	httpClient := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Extract OpenTelemetry context from incoming request
		otelCtx := req.Context()

		// Create a span for the blob plugin
		tracer := trace.SpanFromContext(otelCtx).TracerProvider().Tracer("blob")
		spanCtx, span := tracer.Start(otelCtx, "blob.handle_request",
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

		parts := strings.Split(req.URL.Path, "/")

		// TODO: we should be able to implement this to directly rely on KrakenD instead of writing our own proxy.

		// We encode the MinIO presigned URL to a base64 string in the format:
		// schema://host:port/v1alpha/blob-urls/base64_encoded_presigned_url
		// Here in this plugin, we decode the base64 string to the presigned URL
		// and forward the request to MinIO.
		if len(parts) > 3 && parts[2] == "blob-urls" {
			span.SetAttributes(attribute.String("blob.action", "proxy_presigned_url"))

			// Create a child span for presigned URL processing
			presignedSpanCtx, presignedSpan := tracer.Start(spanCtx, "blob.process_presigned_url",
				trace.WithAttributes(
					attribute.String("blob.encoded_url", parts[3]),
				),
			)
			defer presignedSpan.End()

			blobURLBytes, err := base64.URLEncoding.DecodeString(parts[3])
			if err != nil {
				presignedSpan.RecordError(err)
				presignedSpan.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return
			}

			blobURL, err := url.Parse(string(blobURLBytes))
			if err != nil {
				presignedSpan.RecordError(err)
				presignedSpan.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return
			}

			blobURL.Scheme = "http"
			req.Host = blobURL.Host
			req.URL = blobURL
			req.RequestURI = ""
			req.Header.Del("Authorization")

			presignedSpan.SetAttributes(
				attribute.String("blob.target_host", blobURL.Host),
				attribute.String("blob.target_url", blobURL.String()),
			)

			// Create a child span for the HTTP request to MinIO
			httpSpanCtx, httpSpan := tracer.Start(presignedSpanCtx, "blob.http_minio_request",
				trace.WithAttributes(
					attribute.String("http.target", blobURL.String()),
					attribute.String("http.method", req.Method),
				),
			)
			defer httpSpan.End()

			resp, err := httpClient.Do(req.WithContext(httpSpanCtx))
			if err != nil {
				httpSpan.RecordError(err)
				httpSpan.SetStatus(codes.Error, err.Error())
				presignedSpan.RecordError(err)
				presignedSpan.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				http.Error(w, "Failed to proxy request", http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			httpSpan.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
			presignedSpan.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
			span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

			// Copy response headers
			for k, hs := range resp.Header {
				if k != "Access-Control-Allow-Origin" {
					for _, h := range hs {
						w.Header().Add(k, h)
					}
				}
			}

			// Set status code
			w.WriteHeader(resp.StatusCode)

			// Stream response body
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				presignedSpan.RecordError(err)
				presignedSpan.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				logger.Error(logPrefix, "Failed to stream response body:", err)
			}
			return
		}

		if len(parts) < 6 || parts[1] != "v1alpha" || parts[2] != "namespaces" || parts[4] != "blob-urls" {
			span.SetAttributes(attribute.String("blob.action", "pass_through"))
			h.ServeHTTP(w, req)
			return
		}

		span.SetAttributes(attribute.String("blob.action", "process_blob_request"))
		blobHandler.handler(spanCtx)(w, req)

	}), nil

}

func main() {}

// HandlerRegisterer is the symbol the plugin loader will try to load. It must
// implement the Registerer interface.
var HandlerRegisterer = blob(pluginName)

// This logger will be replaced by the RegisterLogger method to load the one from
// KrakenD.
var logger logging.Logger = logging.NoOp

func (blob) RegisterLogger(v any) {
	l, ok := v.(logging.Logger)
	if !ok {
		return
	}
	logger = l

	logger.Info(logPrefix, "Logger loaded")
}

// The following functions are shortcuts for formatted logging.
var logPrefix = fmt.Sprintf("[PLUGIN: %s]", HandlerRegisterer)

func sanitize(s string) string {
	// html.EscapeString doesn't fully prevent log injection.
	sanitized := strings.ReplaceAll(s, "\n", "")
	sanitized = strings.ReplaceAll(sanitized, "\r", "")
	sanitized = strings.ReplaceAll(sanitized, "\t", "")
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")

	return html.EscapeString(sanitized)

}

func logReq(req *http.Request, v []any) []any {
	logFields := make([]any, 3, len(v)+3)
	logFields[0] = logPrefix
	logFields[1] = req.Method
	logFields[2] = sanitize(req.URL.Path)

	return append(logFields, v...)
}

func logDebug(req *http.Request, v ...any)    { logger.Debug(logReq(req, v)...) }
func logInfo(req *http.Request, v ...any)     { logger.Info(logReq(req, v)...) }
func logWarning(req *http.Request, v ...any)  { logger.Warning(logReq(req, v)...) }
func logError(req *http.Request, v ...any)    { logger.Error(logReq(req, v)...) }
func logCritical(req *http.Request, v ...any) { logger.Critical(logReq(req, v)...) }
func logFatal(req *http.Request, v ...any)    { logger.Fatal(logReq(req, v)...) }
