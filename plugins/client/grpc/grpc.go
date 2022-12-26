package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/trusch/grpc-proxy/proxy"
	"github.com/trusch/grpc-proxy/proxy/codec"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ClientRegisterer is the symbol the plugin loader will try to load. It must implement the RegisterClient interface
var ClientRegisterer = registerer("grpc-access")

type registerer string

func (r registerer) RegisterClients(f func(
	name string,
	handler func(context.Context, map[string]interface{}) (http.Handler, error),
)) {
	f(string(r), r.registerClients)
}

func (r registerer) registerClients(ctx context.Context, extra map[string]interface{}) (http.Handler, error) {
	// Check the passed configuration and initialize the plugin
	name, ok := extra["name"].(string)

	if !ok {
		return nil, errors.New("wrong config")
	}

	if name != string(r) {
		return nil, fmt.Errorf("unknown register %s", name)
	}

	grpcEndpoint, ok := extra["grpc_endpoint"].(string)
	if !ok {
		return nil, errors.New("missing grpc_endpoint in your configuration")
	}

	u, err := url.Parse(grpcEndpoint)
	if err != nil {
		return nil, errors.New("error when parse grpc endpoint")
	}

	director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		// Copy the inbound metadata explicitly.
		outCtx, _ := context.WithCancel(ctx)
		outCtx = metadata.NewOutgoingContext(outCtx, md.Copy())
		if ok {
			conn, err := grpc.DialContext(ctx, u.Host, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.CallContentSubtype((&codec.Proxy{}).Name())))
			return outCtx, conn, err
		}
		return nil, nil, status.Errorf(codes.Unimplemented, "Unknown method")
	}

	s := grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
		grpc.UnknownServiceHandler(proxy.TransparentHandler(director)),
	)

	// Return the actual handler wrapping or your custom logic so it can be used as a replacement for the default http client
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.Proto = "HTTP/2.0"
		req.ProtoMajor = 2
		req.ProtoMinor = 0

		s.ServeHTTP(w, req)
	}), nil
}

func init() {
	fmt.Printf("Plugin: proxy client \"%s\" loaded!!!\n", ClientRegisterer)
}

func main() {}
