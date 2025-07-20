package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/luraproject/lura/v2/logging"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"

	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
)

// pluginName is the plugin name
var pluginName = "multi-auth"

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

	// Create HTTP client with OpenTelemetry instrumentation
	httpClient := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Extract OpenTelemetry context from incoming request
		otelCtx := req.Context()

		// Create a span for the multi-auth plugin
		tracer := trace.SpanFromContext(otelCtx).TracerProvider().Tracer("multi-auth")
		spanCtx, span := tracer.Start(otelCtx, "multi-auth.handle_request",
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

		if req.URL.String() == "/__health" {
			span.SetAttributes(attribute.String("auth.skip_reason", "health_check"))
			h.ServeHTTP(w, req)
		} else if req.URL.String() == "/v1beta/validate_token" {
			span.SetAttributes(attribute.String("auth.skip_reason", "validate_token_endpoint"))
			h.ServeHTTP(w, req)
		} else if req.URL.String() == "/v1beta/auth/login" {
			span.SetAttributes(attribute.String("auth.skip_reason", "login_endpoint"))
			h.ServeHTTP(w, req)
		} else if strings.HasPrefix(authorization, "Basic ") || strings.HasPrefix(authorization, "basic ") {
			// Create a child span for basic auth processing
			basicAuthSpanCtx, basicAuthSpan := tracer.Start(spanCtx, "multi-auth.basic_auth",
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
			grpcSpanCtx, grpcSpan := tracer.Start(basicAuthSpanCtx, "multi-auth.grpc_auth_token_issuer",
				trace.WithAttributes(
					attribute.String("grpc.method", "AuthTokenIssuer"),
					attribute.String("grpc.service", "mgmtPB.MgmtPublicService"),
				),
			)
			defer grpcSpan.End()

			resp, err := mgmtClient.AuthTokenIssuer(grpcSpanCtx, &mgmtPB.AuthTokenIssuerRequest{
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
			grpcSpan.SetAttributes(attribute.String("auth.user_uid", resp.AccessToken.Sub))
			basicAuthSpan.SetAttributes(attribute.String("auth.user_uid", resp.AccessToken.Sub))

			req.Header.Set("Instill-Auth-Type", "user")
			req.Header.Set("Instill-User-Uid", resp.AccessToken.Sub)
			h.ServeHTTP(w, req)

		} else if strings.HasPrefix(authorization, "Bearer instill_sk_") || strings.HasPrefix(authorization, "bearer instill_sk_") {
			// Create a child span for bearer token processing
			bearerSpanCtx, bearerSpan := tracer.Start(spanCtx, "multi-auth.bearer_token",
				trace.WithAttributes(
					attribute.String("auth.type", "bearer_instill_sk"),
				),
			)
			defer bearerSpan.End()

			// Create a child span for the gRPC call
			grpcSpanCtx, grpcSpan := tracer.Start(bearerSpanCtx, "multi-auth.grpc_validate_token",
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

			// Record successful authentication
			grpcSpan.SetAttributes(attribute.String("auth.user_uid", resp.UserUid))
			bearerSpan.SetAttributes(attribute.String("auth.user_uid", resp.UserUid))

			req.Header.Set("Instill-Auth-Type", "user")
			req.Header.Set("Instill-User-Uid", resp.UserUid)
			h.ServeHTTP(w, req)

		} else if authorization == "" {
			// Create a child span for visitor authentication
			_, visitorSpan := tracer.Start(spanCtx, "multi-auth.visitor_auth",
				trace.WithAttributes(
					attribute.String("auth.type", "visitor"),
				),
			)
			defer visitorSpan.End()

			visitorID, _ := uuid.NewV4()
			visitorSpan.SetAttributes(attribute.String("auth.visitor_uid", visitorID.String()))

			req.Header.Set("Instill-Auth-Type", "visitor")
			req.Header.Set("Instill-Visitor-Uid", visitorID.String())
			h.ServeHTTP(w, req)

		} else if req.Header.Get("Accept") == "text/event-stream" {
			// Create a child span for SSE authentication
			sseSpanCtx, sseSpan := tracer.Start(spanCtx, "multi-auth.sse_auth",
				trace.WithAttributes(
					attribute.String("auth.type", "sse"),
				),
			)
			defer sseSpan.End()

			// Currently, KrakenD doesn't support event-stream. To make
			// authentication work, we send a request to the management API
			// first for verification.
			r, err := http.NewRequest("GET", "http://localhost:8080/v1beta/user", nil)
			if err != nil {
				sseSpan.RecordError(err)
				sseSpan.SetStatus(codes.Error, err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			r.Header = req.Header
			r.Header["Accept"][0] = "*/*"

			// Create a child span for the HTTP request
			httpSpanCtx, httpSpan := tracer.Start(sseSpanCtx, "multi-auth.http_user_validation",
				trace.WithAttributes(
					attribute.String("http.target", "http://localhost:8080/v1beta/user"),
					attribute.String("http.method", "GET"),
				),
			)
			defer httpSpan.End()

			resp, err := httpClient.Do(r.WithContext(httpSpanCtx))
			if err != nil {
				httpSpan.RecordError(err)
				httpSpan.SetStatus(codes.Error, err.Error())
				sseSpan.RecordError(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			httpSpan.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

			if resp.StatusCode == 401 {
				httpSpan.SetAttributes(attribute.String("auth.result", "unauthorized"))
				sseSpan.SetAttributes(attribute.String("auth.result", "unauthorized"))
				writeStatusUnauthorized(req, w)
				return
			}

			type user struct {
				User struct {
					UID string `json:"uid"`
				} `json:"user"`
			}
			respBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				httpSpan.RecordError(err)
				sseSpan.RecordError(err)
				writeStatusUnauthorized(req, w)
				return
			}
			defer resp.Body.Close()

			u := user{}
			err = json.Unmarshal(respBytes, &u)
			if err != nil {
				httpSpan.RecordError(err)
				sseSpan.RecordError(err)
				writeStatusUnauthorized(req, w)
				return
			}

			// Record successful authentication
			httpSpan.SetAttributes(attribute.String("auth.user_uid", u.User.UID))
			sseSpan.SetAttributes(attribute.String("auth.user_uid", u.User.UID))

			req.Header.Set("Instill-Auth-Type", "user")
			req.Header.Set("Instill-User-Uid", u.User.UID)
			req.Header.Set("Accept", "text/event-stream")
			h.ServeHTTP(w, req)

		} else {
			// Create a child span for internal routing
			_, internalSpan := tracer.Start(spanCtx, "multi-auth.internal_routing",
				trace.WithAttributes(
					attribute.String("auth.type", "internal"),
				),
			)
			defer internalSpan.End()

			req.Header.Set("Instill-Auth-Type", "user")
			req.URL.Path = "/internal" + req.URL.Path
			h.ServeHTTP(w, req)
		}
	}), nil
}

func main() {}

// This logger is replaced by the RegisterLogger method to load the one from KrakenD
var logger = logging.NoOp

func (registerer) RegisterLogger(v any) {
	l, ok := v.(logging.BasicLogger)
	if !ok {
		return
	}
	logger = l
	logger.Info(fmt.Sprintf("[PLUGIN: %s] Logger loaded", HandlerRegisterer))
}
