module registry-plugin

go 1.24.4

require (
	github.com/distribution/distribution v2.8.3+incompatible
	github.com/frankban/quicktest v1.14.6
	github.com/instill-ai/protogen-go v0.3.3-alpha.0.20241024152819-5ed9f53b5c8a
	github.com/luraproject/lura/v2 v2.10.0
	go.opentelemetry.io/otel v1.33.0
	go.opentelemetry.io/otel/trace v1.33.0
	google.golang.org/grpc v1.66.0
)

require (
	cloud.google.com/go/longrunning v0.5.12 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240812133136-8ffd90a71988 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240812133136-8ffd90a71988 // indirect
	google.golang.org/protobuf v1.36.3 // indirect
)

replace golang.org/x/net => golang.org/x/net v0.38.0

replace golang.org/x/sys => golang.org/x/sys v0.31.0

replace golang.org/x/text => golang.org/x/text v0.23.0
