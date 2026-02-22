package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"

	mgmtPB "github.com/instill-ai/protogen-go/mgmt/v1beta"
)

// pluginName is the plugin name
var pluginName = "simple-auth"

// HandlerRegisterer is the symbol the plugin loader will try to load. It must implement the Registerer interface
var HandlerRegisterer = registerer(pluginName)

type registerer string

func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]any, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

func writeStatusUnauthorized(req *http.Request, w http.ResponseWriter) {
	if req.ProtoMajor == 2 && strings.Contains(req.Header.Get("Content-Type"), "application/grpc") {
		w.Header().Set("Content-Type", "application/grpc")
		w.Header().Set("Trailer", "Grpc-Status")
		w.Header().Add("Trailer", "Grpc-Message")
		w.Header().Set("Grpc-Status", "16")               // UNAUTHENTICATED
		w.Header().Set("Grpc-Message", "Unauthenticated") // UNAUTHENTICATED
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
	}
}

func (r registerer) registerHandlers(ctx context.Context, extra map[string]any, h http.Handler) (http.Handler, error) {
	config, ok := extra[pluginName].(map[string]any)
	if !ok {
		return h, errors.New("configuration not found")
	}

	mgmtClient, _ := InitMgmtPublicServiceClient(context.Background(), config["grpc_server"].(string), "", "")

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Strip all gateway-internal headers to prevent external callers from
		// spoofing identity by injecting these headers directly.
		req.Header.Del("Instill-Auth-Type")
		req.Header.Del("Instill-User-Uid")
		req.Header.Del("Instill-Visitor-Uid")
		req.Header.Del("Instill-Namespace-Id")
		req.Header.Del("Instill-Internal-Request-Uid")

		// Extract OpenTelemetry context from incoming request
		otelCtx := req.Context()

		// Create a span for the simple-auth plugin
		tracer := trace.SpanFromContext(otelCtx).TracerProvider().Tracer("simple-auth")
		spanCtx, span := tracer.Start(otelCtx, "simple-auth.handle_request",
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

		authorization := req.Header.Get("Authorization")
		urlPath := req.URL.Path

		// Public endpoints that don't require authentication
		isPublicEndpoint := urlPath == "/__health" ||
			urlPath == "/v1beta/validate-token" ||
			urlPath == "/v1beta/auth/login" ||
			strings.Contains(urlPath, "/health/") ||
			strings.Contains(urlPath, "/ready/") ||
			strings.HasSuffix(urlPath, "/Liveness") ||
			strings.HasSuffix(urlPath, "/Readiness")

		if isPublicEndpoint {
			span.SetAttributes(attribute.String("auth.skip_reason", "public_endpoint"))
			span.SetAttributes(attribute.String("auth.public_path", urlPath))
			h.ServeHTTP(w, req)
		} else if strings.HasPrefix(authorization, "Basic ") || strings.HasPrefix(authorization, "basic ") {
			// Create a child span for basic auth processing
			basicAuthSpanCtx, basicAuthSpan := tracer.Start(spanCtx, "simple-auth.basic_auth",
				trace.WithAttributes(
					attribute.String("auth.type", "basic"),
				),
			)
			defer basicAuthSpan.End()

			basicAuth := strings.Split(authorization, " ")[1]

			basicAuthDecoded, err := base64.StdEncoding.DecodeString(basicAuth)
			if err != nil {
				basicAuthSpan.RecordError(err)
				basicAuthSpan.SetStatus(codes.Error, "Failed to decode basic auth")
				writeStatusUnauthorized(req, w)
				return
			}

			// Create a child span for the gRPC call
			grpcSpanCtx, grpcSpan := tracer.Start(basicAuthSpanCtx, "simple-auth.grpc_authenticate_user",
				trace.WithAttributes(
					attribute.String("grpc.method", "AuthenticateUser"),
					attribute.String("grpc.service", "mgmtPB.MgmtPublicService"),
				),
			)
			defer grpcSpan.End()

			resp, err := mgmtClient.AuthenticateUser(grpcSpanCtx, &mgmtPB.AuthenticateUserRequest{
				Username: strings.Split(string(basicAuthDecoded), ":")[0],
				Password: strings.Split(string(basicAuthDecoded), ":")[1],
			})
			if err != nil {
				grpcSpan.RecordError(err)
				grpcSpan.SetStatus(codes.Error, err.Error())
				basicAuthSpan.RecordError(err)
				writeStatusUnauthorized(req, w)
				return
			}

			// Record successful authentication
			grpcSpan.SetAttributes(attribute.String("auth.user_uid", resp.UserUid))
			basicAuthSpan.SetAttributes(attribute.String("auth.user_uid", resp.UserUid))

			req.Header.Set("Instill-Auth-Type", "user")
			req.Header.Set("Instill-User-Uid", resp.UserUid)
			h.ServeHTTP(w, req)

		} else if strings.HasPrefix(authorization, "Bearer instill_sk_") || strings.HasPrefix(authorization, "bearer instill_sk_") {
			// API key validation via mgmt-backend
			bearerSpanCtx, bearerSpan := tracer.Start(spanCtx, "simple-auth.bearer_token",
				trace.WithAttributes(
					attribute.String("auth.type", "bearer_api_key"),
				),
			)
			defer bearerSpan.End()

			grpcSpanCtx, grpcSpan := tracer.Start(bearerSpanCtx, "simple-auth.grpc_validate_token",
				trace.WithAttributes(
					attribute.String("grpc.method", "ValidateToken"),
					attribute.String("grpc.service", "mgmtPB.MgmtPublicService"),
				),
			)
			defer grpcSpan.End()

			ctx = metadata.AppendToOutgoingContext(grpcSpanCtx, "Authorization", req.Header.Get("authorization"))
			resp, err := mgmtClient.ValidateToken(ctx, &mgmtPB.ValidateTokenRequest{})
			if err != nil {
				grpcSpan.RecordError(err)
				grpcSpan.SetStatus(codes.Error, err.Error())
				bearerSpan.RecordError(err)
				writeStatusUnauthorized(req, w)
				return
			}

			userUID := resp.User
			if after, ok := strings.CutPrefix(resp.User, "users/"); ok {
				userUID = after
			}
			grpcSpan.SetAttributes(attribute.String("auth.user_uid", userUID))
			bearerSpan.SetAttributes(attribute.String("auth.user_uid", userUID))

			req.Header.Set("Instill-Auth-Type", "user")
			req.Header.Set("Instill-User-Uid", userUID)
			h.ServeHTTP(w, req)

		} else {
			// Unknown or missing authorization - return unauthorized
			// CE edition only supports Basic Auth and Bearer instill_sk_ tokens
			span.SetAttributes(attribute.String("auth.result", "unauthorized"))
			span.SetAttributes(attribute.String("auth.reason", "unsupported_auth_type"))
			writeStatusUnauthorized(req, w)
		}
	}), nil
}

func main() {}

// Logger is the interface for the logger (matches KrakenD's expected interface)
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

// This logger is replaced by the RegisterLogger method to load the one from KrakenD
var logger Logger = noopLogger{}

func (registerer) RegisterLogger(v any) {
	l, ok := v.(Logger)
	if !ok {
		return
	}
	logger = l
	logger.Info(fmt.Sprintf("[PLUGIN: %s] Logger loaded", HandlerRegisterer))
}
