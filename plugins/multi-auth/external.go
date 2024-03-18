package main

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
)

const MaxPayloadSize = 1024 * 1024 * 32

func InitMgmtPublicServiceClient(ctx context.Context, server string, cert string, key string) (mgmtPB.MgmtPublicServiceClient, *grpc.ClientConn) {

	var clientDialOpts grpc.DialOption
	if cert != "" && key != "" {
		creds, err := credentials.NewServerTLSFromFile(cert, key)
		if err != nil {
			panic(err)
		}
		clientDialOpts = grpc.WithTransportCredentials(creds)
	} else {
		clientDialOpts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	clientConn, err := grpc.Dial(server, clientDialOpts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(MaxPayloadSize), grpc.MaxCallSendMsgSize(MaxPayloadSize)))
	if err != nil {
		return nil, nil
	}

	return mgmtPB.NewMgmtPublicServiceClient(clientConn), clientConn
}
