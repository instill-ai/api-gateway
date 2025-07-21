module grpc_proxy_plugin

go 1.24.4

require (
	github.com/luraproject/lura/v2 v2.10.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.53.0
	go.opentelemetry.io/otel v1.33.0
	golang.org/x/net v0.38.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.33.0 // indirect
	go.opentelemetry.io/otel/trace v1.33.0 // indirect
	golang.org/x/text v0.23.0 // indirect
)

replace go.opentelemetry.io/otel/metric => go.opentelemetry.io/otel/metric v1.33.0

replace go.opentelemetry.io/otel/sdk => go.opentelemetry.io/otel/sdk v1.28.0

replace go.opentelemetry.io/otel/sdk/metric => go.opentelemetry.io/otel/sdk/metric v1.28.0

replace go.opentelemetry.io/otel/trace => go.opentelemetry.io/otel/trace v1.33.0

replace golang.org/x/sys => golang.org/x/sys v0.31.0

replace github.com/go-logr/logr => github.com/go-logr/logr v1.4.2
