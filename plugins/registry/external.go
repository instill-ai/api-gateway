package main

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
)

const maxPayloadSize = 1024 * 1024 * 32

func initMgmtPublicServiceClient(_ context.Context, server string, cert string, key string) (mgmtPB.MgmtPublicServiceClient, *grpc.ClientConn) {
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

	clientConn, err := grpc.Dial(server, clientDialOpts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxPayloadSize), grpc.MaxCallSendMsgSize(maxPayloadSize)))
	if err != nil {
		return nil, nil
	}

	return mgmtPB.NewMgmtPublicServiceClient(clientConn), clientConn
}

func initMgmtPrivateServiceClient(_ context.Context, server string, cert string, key string) (mgmtPB.MgmtPrivateServiceClient, *grpc.ClientConn) {
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

	clientConn, err := grpc.Dial(server, clientDialOpts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxPayloadSize), grpc.MaxCallSendMsgSize(maxPayloadSize)))
	if err != nil {
		return nil, nil
	}

	return mgmtPB.NewMgmtPrivateServiceClient(clientConn), clientConn
}
