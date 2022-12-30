package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/akosyakov/grpc-proxy/proxy"
	"github.com/gogo/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"

	"plugin/internal/logger"
)

// pluginName is the plugin name
var pluginName = "grpc-proxy"

// HandlerRegisterer is the symbol the plugin loader will try to load. It must implement the Registerer interface
var HandlerRegisterer = registerer(pluginName)

type registerer string

func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

func (r registerer) registerHandlers(ctx context.Context, extra map[string]interface{}, h http.Handler) (http.Handler, error) {
	
	logger, _ := logger.GetZapLogger()
	defer func() {
		// can't handle the error due to https://github.com/uber-go/zap/issues/880
		_ = logger.Sync()
	}()
	grpc_zap.ReplaceGrpcLoggerV2(logger)

	config, ok := extra[pluginName].(map[string]interface{})
	if !ok {
		return h, errors.New("configuration not found")
	}

	grpcServerOpts := []grpc.ServerOption{
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			streamAppendMetadataInterceptor,
			grpc_zap.StreamServerInterceptor(logger),
			grpc_recovery.StreamServerInterceptor(recoveryInterceptorOpt()),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			unaryAppendMetadataInterceptor,
			grpc_zap.UnaryServerInterceptor(logger),
			grpc_recovery.UnaryServerInterceptor(recoveryInterceptorOpt()),
		)),
	}

	// Register gRPC servers
	grpcServers := map[string]*grpc.Server{}
	for srv, endpoint := range config {

		target, _ := url.Parse(endpoint.(string))

		conn, err := grpc.Dial(		
			target.Host,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithCodec(proxy.Codec()))
	
		director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
			md, ok := metadata.FromIncomingContext(ctx)
			outCtx := metadata.NewOutgoingContext(ctx, md.Copy())
			if ok {			
				return outCtx, conn, err
			}
			return nil, nil, status.Errorf(codes.Unimplemented, "Unknown method")
		}
	
		grpcServerOpts = append(
			grpcServerOpts,	
			grpc.UnknownServiceHandler(proxy.TransparentHandler(director)))
	
		grpcServers[srv] = grpc.NewServer(grpcServerOpts...)
	}

	// Return the actual handler wrapping or your custom logic so it can be used as a replacement for the default http handler
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// According to https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md#requests
		// The Content-Type of gRPC has the "application/grpc" prefix
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {								
			// r.URL.Path is /vdp.{service_name}.v1alpha.{service_name}Service/{method}			
			srv := strings.Split(strings.Split(r.URL.Path, "/")[1], ".")[1]
			grpcServers[srv].ServeHTTP(w, r)
		} else {
			h.ServeHTTP(w, r)
		}
	}), nil
}

func removeDuplicateHeader(strSlice []string) []string {
    allKeys := make(map[string]bool)
    list := []string{}
    for _, item := range strSlice {
        if _, value := allKeys[item]; !value {
            allKeys[item] = true
            list = append(list, item)
        }
    }
    return list
}

func init() {
	fmt.Printf("Plugin: router handler \"%s\" loaded!!!\n", HandlerRegisterer)
}

func main() {}
