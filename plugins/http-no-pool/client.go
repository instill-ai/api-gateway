package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/luraproject/lura/v2/logging"
)

// ClientRegisterer is the symbol the plugin loader will try to load.
var ClientRegisterer = clientRegisterer("http-no-pool-client")

type clientRegisterer string

func (r clientRegisterer) RegisterClients(f func(
	name string,
	handler func(context.Context, map[string]any) (http.Handler, error),
)) {
	f(string(r), r.registerClients)
}

func (r clientRegisterer) registerClients(_ context.Context, extra map[string]any) (http.Handler, error) {
	name, ok := extra["name"].(string)
	if !ok {
		return nil, errors.New("wrong config")
	}
	if name != string(r) {
		return nil, fmt.Errorf("unknown register %s", name)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Per-request transport prevents HTTP/1.1 connection pooling from
		// pinning traffic to a single backend pod. With a K8s Service, each
		// new TCP connection triggers a fresh kube-proxy routing decision,
		// distributing load across all ready replicas — including those
		// added by HPA autoscaling.
		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.DisableKeepAlives = true
		otelTransport := otelhttp.NewTransport(tr)
		httpClient := http.Client{Transport: otelTransport}
		defer httpClient.CloseIdleConnections()

		otelCtx := req.Context()
		tracer := trace.SpanFromContext(otelCtx).TracerProvider().Tracer("http-no-pool-client")
		spanCtx, span := tracer.Start(otelCtx, "http-no-pool.client.request",
			trace.WithAttributes(
				attribute.String("http.method", req.Method),
				attribute.String("http.url", req.URL.String()),
				attribute.String("plugin.name", "http-no-pool-client"),
			),
		)
		defer span.End()

		newReq := req.WithContext(spanCtx)
		otel.GetTextMapPropagator().Inject(spanCtx, propagation.HeaderCarrier(newReq.Header))

		httpSpanCtx, httpSpan := tracer.Start(spanCtx, "http-no-pool.client.http_request",
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

		span.SetAttributes(attribute.Int("http.response_size", len(respBytes)))

		for k, hs := range resp.Header {
			for _, h := range hs {
				w.Header().Add(k, h)
			}
		}

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
}
