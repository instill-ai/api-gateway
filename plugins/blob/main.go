package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

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

type cacheEntry struct {
	presignedURL string
	expiresAt    time.Time
}

var (
	blobCacheMu sync.RWMutex
	blobCache   = make(map[string]cacheEntry)
)

const cacheTTL = 5 * time.Minute

func getCachedPresignedURL(objectID string) (string, bool) {
	blobCacheMu.RLock()
	defer blobCacheMu.RUnlock()

	entry, ok := blobCache[objectID]
	if !ok || time.Now().After(entry.expiresAt) {
		return "", false
	}
	return entry.presignedURL, true
}

func setCachedPresignedURL(objectID, presignedURL string) {
	blobCacheMu.Lock()
	defer blobCacheMu.Unlock()

	blobCache[objectID] = cacheEntry{
		presignedURL: presignedURL,
		expiresAt:    time.Now().Add(cacheTTL),
	}
}

// resolveObjectID calls the artifact-backend internal endpoint to resolve an
// object_id into a fresh presigned URL.
func resolveObjectID(ctx context.Context, httpClient *http.Client, artifactHost, objectID string) (string, error) {
	reqURL := fmt.Sprintf("http://%s/internal/resolve-blob/%s", artifactHost, objectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating resolve request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling artifact-backend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("artifact-backend returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	return strings.TrimSpace(string(body)), nil
}

// tryDecodeAsPresignedURL attempts to base64-decode the segment and parse it
// as a URL. Returns the parsed URL if successful, nil otherwise.
func tryDecodeAsPresignedURL(encoded string) *url.URL {
	blobURLBytes, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		blobURLBytes, err = base64.RawURLEncoding.DecodeString(encoded)
		if err != nil {
			return nil
		}
	}

	parsed, err := url.Parse(string(blobURLBytes))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil
	}

	return parsed
}

func (r blob) registerHandlers(_ context.Context, extra map[string]any, h http.Handler) (http.Handler, error) {
	cfg, ok := extra[pluginName].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("configuration not found")
	}

	artifactPrivateHostport, _ := cfg["artifact_private_hostport"].(string)

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		MaxConnsPerHost:     50,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  true,
		ForceAttemptHTTP2:   false,
	}
	httpClient := http.Client{Transport: otelhttp.NewTransport(transport)}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		otelCtx := req.Context()

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

		req = req.WithContext(spanCtx)

		parts := strings.Split(req.URL.Path, "/")

		if len(parts) > 3 && parts[2] == "blob-urls" {
			encodedOrID := parts[3]

			// Mode 1: Try to decode as base64-encoded presigned URL
			if blobURL := tryDecodeAsPresignedURL(encodedOrID); blobURL != nil {
				span.SetAttributes(attribute.String("blob.action", "proxy_presigned_url"))
				proxyToMinIO(spanCtx, tracer, &httpClient, w, req, blobURL, false)
				return
			}

			// Mode 2: Treat as object_id, resolve via artifact-backend
			if artifactPrivateHostport == "" {
				span.RecordError(fmt.Errorf("artifact_private_hostport not configured"))
				span.SetStatus(codes.Error, "artifact_private_hostport not configured")
				http.Error(w, "blob resolution not available", http.StatusServiceUnavailable)
				return
			}

			span.SetAttributes(
				attribute.String("blob.action", "resolve_object_id"),
				attribute.String("blob.object_id", encodedOrID),
			)

			// Check in-memory cache first
			presignedURLStr, cached := getCachedPresignedURL(encodedOrID)
			if !cached {
				var err error
				presignedURLStr, err = resolveObjectID(spanCtx, &httpClient, artifactPrivateHostport, encodedOrID)
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					http.Error(w, "Failed to resolve object", http.StatusNotFound)
					return
				}
				setCachedPresignedURL(encodedOrID, presignedURLStr)
			}
			span.SetAttributes(attribute.Bool("blob.cache_hit", cached))

			blobURL, err := url.Parse(presignedURLStr)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				http.Error(w, "Failed to parse resolved URL", http.StatusInternalServerError)
				return
			}
			blobURL.Scheme = "http"

			proxyToMinIO(spanCtx, tracer, &httpClient, w, req, blobURL, true)
			return
		}

		span.SetAttributes(attribute.String("blob.action", "pass_through"))
		h.ServeHTTP(w, req)

	}), nil

}

// proxyToMinIO proxies the request to MinIO using the resolved presigned URL.
// When isObjectID is true, Cache-Control headers are added for browser caching.
func proxyToMinIO(
	ctx context.Context,
	tracer trace.Tracer,
	httpClient *http.Client,
	w http.ResponseWriter,
	req *http.Request,
	blobURL *url.URL,
	isObjectID bool,
) {
	_, httpSpan := tracer.Start(ctx, "blob.http_minio_request",
		trace.WithAttributes(
			attribute.String("http.target", blobURL.String()),
			attribute.String("http.method", req.Method),
		),
	)
	defer httpSpan.End()

	newReq, err := http.NewRequestWithContext(ctx, req.Method, blobURL.String(), req.Body)
	if err != nil {
		httpSpan.RecordError(err)
		httpSpan.SetStatus(codes.Error, err.Error())
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	if req.ContentLength > 0 {
		newReq.ContentLength = req.ContentLength
	}

	if accept := req.Header.Get("Accept"); accept != "" {
		newReq.Header.Set("Accept", accept)
	}
	if acceptEncoding := req.Header.Get("Accept-Encoding"); acceptEncoding != "" {
		newReq.Header.Set("Accept-Encoding", acceptEncoding)
	}
	if contentType := req.Header.Get("Content-Type"); contentType != "" {
		newReq.Header.Set("Content-Type", contentType)
	}

	resp, err := httpClient.Do(newReq)
	if err != nil {
		httpSpan.RecordError(err)
		httpSpan.SetStatus(codes.Error, err.Error())
		http.Error(w, "Failed to proxy request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	httpSpan.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	for k, hs := range resp.Header {
		if k != "Access-Control-Allow-Origin" {
			for _, h := range hs {
				w.Header().Add(k, h)
			}
		}
	}

	// For object_id-resolved requests, add aggressive browser caching since
	// objects are immutable. Legacy base64 presigned URLs manage their own
	// expiry via the signed URL parameters.
	if isObjectID {
		w.Header().Set("Cache-Control", "public, max-age=86400")
	}

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		if isClientDisconnect(err) {
			httpSpan.SetAttributes(attribute.Bool("blob.client_disconnected", true))
			logger.Debug(logPrefix, "Client disconnected during streaming:", err)
		} else {
			httpSpan.RecordError(err)
			httpSpan.SetStatus(codes.Error, err.Error())
			logger.Error(logPrefix, "Failed to stream response body:", err)
		}
	}
}

// isClientDisconnect returns true when the error indicates the downstream
// client (or an intermediate proxy such as the GCP LB) closed the connection.
// These are expected during normal browsing (e.g. user navigates away while a
// blob is streaming) and should not be logged at ERROR level.
func isClientDisconnect(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()
	if strings.Contains(msg, "http2: stream closed") ||
		strings.Contains(msg, "client disconnected") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "connection reset by peer") {
		return true
	}

	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return true
	}

	return errors.Is(err, context.Canceled)
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
