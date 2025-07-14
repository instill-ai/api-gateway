package main

import (
	"context"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
)

const MaxPayloadSize = 1024 * 1024 * 32

func InitMgmtPublicServiceClient(ctx context.Context, server string, cert string, key string) (mgmtPB.MgmtPublicServiceClient, *grpc.ClientConn) {

	var dialOpts []grpc.DialOption

	dialOpts = append(dialOpts,
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(MaxPayloadSize),
			grpc.MaxCallSendMsgSize(MaxPayloadSize),
		),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)

	if cert != "" && key != "" {
		creds, err := credentials.NewServerTLSFromFile(cert, key)
		if err != nil {
			panic(err)
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	clientConn, err := grpc.NewClient(server, dialOpts...)
	if err != nil {
		return nil, nil
	}

	return mgmtPB.NewMgmtPublicServiceClient(clientConn), clientConn
}
