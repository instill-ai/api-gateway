module sse_streaming_plugin

go 1.24.4

require (
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.53.0
	go.opentelemetry.io/otel v1.33.0
	go.opentelemetry.io/otel/trace v1.33.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.33.0 // indirect
)

replace github.com/go-logr/logr => github.com/go-logr/logr v1.4.2

replace go.opentelemetry.io/otel/metric => go.opentelemetry.io/otel/metric v1.33.0
