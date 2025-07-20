package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/luraproject/lura/v2/logging"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/http2"
)

// ClientRegisterer is the symbol the plugin loader will try to load. It must implement the RegisterClient interface
var ClientRegisterer = clientRegisterer("grpc-proxy-client")

type clientRegisterer string

func (r clientRegisterer) RegisterClients(f func(
	name string,
	handler func(context.Context, map[string]any) (http.Handler, error),
)) {
	f(string(r), r.registerClients)
}

func (r clientRegisterer) registerClients(_ context.Context, extra map[string]any) (http.Handler, error) {
	// check the passed configuration and initialize the plugin
	name, ok := extra["name"].(string)
	if !ok {
		return nil, errors.New("wrong config")
	}
	if name != string(r) {
		return nil, fmt.Errorf("unknown register %s", name)
	}

	// Create HTTP client with OpenTelemetry instrumentation
	tr := &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(netw, addr)
		},
	}

	// Wrap the transport with OpenTelemetry instrumentation
	otelTransport := otelhttp.NewTransport(tr)
	httpClient := http.Client{Transport: otelTransport}
	defer httpClient.CloseIdleConnections()

	// return the actual handler wrapping or your custom logic so it can be used as a replacement for the default http handler
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Extract OpenTelemetry context from the request
		otelCtx := req.Context()

		// Create a span for the gRPC proxy client
		tracer := trace.SpanFromContext(otelCtx).TracerProvider().Tracer("grpc-proxy-client")
		spanCtx, span := tracer.Start(otelCtx, "grpc-proxy.client.request",
			trace.WithAttributes(
				attribute.String("http.method", req.Method),
				attribute.String("http.url", req.URL.String()),
				attribute.String("plugin.name", "grpc-proxy-client"),
			),
		)
		defer span.End()

		req.Header.Set("content-type", "application/grpc")

		// Create a new request with the OpenTelemetry context
		newReq := req.WithContext(spanCtx)

		// Propagate OpenTelemetry context to headers
		otel.GetTextMapPropagator().Inject(spanCtx, propagation.HeaderCarrier(newReq.Header))

		// Create a child span for the actual HTTP request
		httpSpanCtx, httpSpan := tracer.Start(spanCtx, "grpc-proxy.client.http_request",
			trace.WithAttributes(
				attribute.String("http.target", newReq.URL.String()),
				attribute.String("http.scheme", newReq.URL.Scheme),
				attribute.String("http.host", newReq.URL.Host),
			),
		)
		defer httpSpan.End()

		resp, err := httpClient.Do(newReq.WithContext(httpSpanCtx))
		if err != nil {
			httpSpan.RecordError(err)
			httpSpan.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Record response status in span
		httpSpan.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
		span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			httpSpan.RecordError(err)
			span.RecordError(err)
			logger.Warning(err.Error())
		}
		defer resp.Body.Close()

		if resp.Body == nil {
			return
		}

		// Record response size in span
		span.SetAttributes(attribute.Int("http.response_size", len(respBytes)))

		for k, hs := range resp.Header {
			for _, h := range hs {
				w.Header().Add(k, h)
			}
		}

		// We can only get Trailer after reading the body
		for k, hs := range resp.Trailer {
			for _, h := range hs {
				w.Header().Add(k, h)
			}
		}

		w.WriteHeader(resp.StatusCode)
		w.Write(respBytes)
	}), nil
}

func (clientRegisterer) RegisterLogger(v any) {
	l, ok := v.(logging.BasicLogger)
	if !ok {
		return
	}
	logger = l
	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Logger loaded", ClientRegisterer))
}
