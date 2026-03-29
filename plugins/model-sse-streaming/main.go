package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var pluginName = "model-sse-streaming"

var HandlerRegisterer = registerer(pluginName)

type registerer string

func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]any, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

func (r registerer) registerHandlers(_ context.Context, extra map[string]any, h http.Handler) (http.Handler, error) {
	config, ok := extra[pluginName].(map[string]any)
	if !ok {
		return h, errors.New("configuration not found")
	}

	backendHostport, ok := config["backend_hostport"].(string)
	if !ok || backendHostport == "" {
		return h, errors.New("backend_hostport is required")
	}

	if !strings.Contains(backendHostport, ":") {
		return h, errors.New("invalid backend_hostport format (expected host:port)")
	}

	useTLS, _ := config["tls"].(bool)
	selfURL := "http://localhost:8080/v1beta/user"
	if useTLS {
		selfURL = "https://localhost:8080/v1beta/user"
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: useTLS} // #nosec G402

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		path := strings.TrimSuffix(req.URL.Path, "/")

		isOpenAI := req.Method == "POST" && path == "/v1/chat/completions"
		isAnthropic := req.Method == "POST" && path == "/v1/messages"
		isModels := req.Method == "GET" && path == "/v1/models"

		if !isOpenAI && !isAnthropic && !isModels {
			h.ServeHTTP(w, req)
			return
		}

		ctx, span := otel.Tracer(pluginName).Start(req.Context(), "handleRequest")
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", req.Method),
			attribute.String("http.url", req.URL.String()),
		)

		httpClient := http.Client{Transport: transport}

		existingAuthType := req.Header.Get("Instill-Auth-Type")
		existingUserUID := req.Header.Get("Instill-User-Uid")

		if existingAuthType != "" && existingUserUID != "" {
			span.SetAttributes(attribute.String("auth.source", "internal_headers"))
		} else {
			userUID, err := validateToken(ctx, req, httpClient, selfURL)
			if err != nil {
				span.SetStatus(codes.Error, "auth validation failed")
				span.RecordError(err)
				http.Error(w, `{"error":{"message":"Unauthorized: Access token is missing or invalid","type":"authentication_error","code":"unauthorized"}}`, http.StatusUnauthorized)
				return
			}
			req.Header.Set("Instill-User-Uid", userUID)
			req.Header.Set("Instill-Requester-Uid", userUID)
			req.Header.Set("Instill-Auth-Type", "user")
			span.SetAttributes(attribute.String("auth.source", "jwt_validation"))
		}

		if isModels {
			proxyJSON(ctx, w, req, httpClient, "GET", backendHostport)
		} else {
			proxySSE(ctx, w, req, httpClient, backendHostport)
		}

	}), nil
}

func validateToken(ctx context.Context, req *http.Request, httpClient http.Client, selfURL string) (string, error) {
	ctx, span := otel.Tracer(pluginName).Start(ctx, "validateToken")
	defer span.End()

	authHeader := req.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		if apiKey := req.Header.Get("X-Api-Key"); apiKey != "" {
			authHeader = "Bearer " + apiKey
			req.Header.Set("Authorization", authHeader)
		} else {
			return "", errors.New("invalid authorization header")
		}
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	r, err := http.NewRequestWithContext(ctx, "GET", selfURL, nil)
	if err != nil {
		return "", err
	}
	r.Header.Set("Authorization", authHeader)
	r.Header.Set("Accept", "*/*")

	resp, err := httpClient.Do(r)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return "", errors.New("unauthorized")
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", errors.New("invalid JWT format")
	}

	decoded, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	var claims struct {
		InstillUserUID string `json:"instill_user_uid"`
	}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return "", err
	}
	if claims.InstillUserUID == "" {
		return "", errors.New("instill_user_uid not found in JWT claims")
	}

	return claims.InstillUserUID, nil
}

func proxySSE(ctx context.Context, w http.ResponseWriter, r *http.Request, httpClient http.Client, backendHostport string) {
	_, span := otel.Tracer(pluginName).Start(ctx, "proxySSE")
	defer span.End()

	targetURL := fmt.Sprintf("http://%s%s", backendHostport, r.URL.Path)
	span.SetAttributes(attribute.String("proxy.target_url", targetURL))

	proxyReq, err := http.NewRequestWithContext(ctx, r.Method, targetURL, r.Body)
	if err != nil {
		span.SetStatus(codes.Error, "failed to create request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	proxyReq.Header = r.Header

	resp, err := httpClient.Do(proxyReq)
	if err != nil {
		span.SetStatus(codes.Error, "backend request failed")
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	isSSE := strings.HasPrefix(ct, "text/event-stream")

	if !isSSE {
		for k, v := range resp.Header {
			w.Header()[k] = v
		}
		w.WriteHeader(resp.StatusCode)
		if _, err := bufio.NewReader(resp.Body).WriteTo(w); err != nil {
			span.RecordError(err)
		}
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher := findFlushableWriter(w)

	reader := bufio.NewReader(resp.Body)
	eventCount := 0

	for {
		dataCh := make(chan string, 1)
		errCh := make(chan error, 1)

		go func() {
			line, err := reader.ReadString('\n')
			if err != nil {
				errCh <- err
				return
			}
			dataCh <- line
		}()

		select {
		case <-r.Context().Done():
			span.SetAttributes(attribute.Int("events.streamed", eventCount))
			return
		case line := <-dataCh:
			eventCount++
			if _, err := w.Write([]byte(line)); err != nil {
				return
			}
			if flusher != nil {
				flusher.Flush()
			}
		case <-errCh:
			span.SetAttributes(attribute.Int("events.streamed", eventCount))
			return
		case <-time.After(600 * time.Second):
			span.SetAttributes(attribute.Int("events.streamed", eventCount))
			return
		}
	}
}

func proxyJSON(ctx context.Context, w http.ResponseWriter, r *http.Request, httpClient http.Client, method string, backendHostport string) {
	_, span := otel.Tracer(pluginName).Start(ctx, "proxyJSON")
	defer span.End()

	targetURL := fmt.Sprintf("http://%s%s", backendHostport, r.URL.Path)
	proxyReq, err := http.NewRequestWithContext(ctx, method, targetURL, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	proxyReq.Header = r.Header

	resp, err := httpClient.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	if _, err := bufio.NewReader(resp.Body).WriteTo(w); err != nil {
		span.RecordError(err)
	}
}

// findFlushableWriter extracts an http.Flusher from the ResponseWriter chain.
// KrakenD wraps the ResponseWriter which hides the Flusher interface.
func findFlushableWriter(w http.ResponseWriter) http.Flusher {
	if f, ok := w.(http.Flusher); ok {
		return f
	}

	rwType := reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()
	current := w

	for depth := 0; depth < 10; depth++ {
		val := reflect.ValueOf(current)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		if val.Kind() != reflect.Struct {
			break
		}

		var nextWriter http.ResponseWriter
		for j := 0; j < val.NumField(); j++ {
			fieldType := val.Type().Field(j)
			field := val.Field(j)

			if !fieldType.Type.Implements(rwType) && fieldType.Type != rwType {
				continue
			}

			var innerWriter http.ResponseWriter
			if field.CanInterface() {
				if iw, ok := field.Interface().(http.ResponseWriter); ok {
					innerWriter = iw
				}
			} else {
				fieldPtr := unsafe.Pointer(field.UnsafeAddr()) // #nosec G103
				innerWriter = *(*http.ResponseWriter)(fieldPtr)
			}

			if innerWriter == nil {
				continue
			}

			if f, ok := innerWriter.(http.Flusher); ok {
				return f
			}
			nextWriter = innerWriter
		}

		if nextWriter == nil {
			break
		}
		current = nextWriter
	}

	return nil
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
