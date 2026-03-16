module registry-plugin

go 1.25.6

require (
	github.com/distribution/distribution v2.8.3+incompatible
	github.com/frankban/quicktest v1.14.6
	github.com/instill-ai/protogen-go v0.3.3-alpha.0.20260216034810-ff9e6f04b974
	github.com/luraproject/lura/v2 v2.12.1
	go.opentelemetry.io/otel v1.34.0
	go.opentelemetry.io/otel/trace v1.34.0
	google.golang.org/grpc v1.71.0
)

require (
	cloud.google.com/go/longrunning v0.5.12 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250106144421-5f5ef82da422 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250922171735-9219d122eba9 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
)

replace golang.org/x/net => golang.org/x/net v0.47.0

replace golang.org/x/sys => golang.org/x/sys v0.38.0

replace golang.org/x/text => golang.org/x/text v0.31.0

replace google.golang.org/protobuf => google.golang.org/protobuf v1.36.10

replace google.golang.org/genproto/googleapis/rpc => google.golang.org/genproto/googleapis/rpc v0.0.0-20251002232023-7c0ddcbb5797

replace google.golang.org/genproto/googleapis/api => google.golang.org/genproto/googleapis/api v0.0.0-20251002232023-7c0ddcbb5797
