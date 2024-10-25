package main

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const maxPayloadSize = 1024 * 1024 * 32

func newGRPCConn(server, cert, key string) (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials
	if cert == "" || key == "" {
		creds = insecure.NewCredentials()
	} else {
		var err error
		creds, err = credentials.NewServerTLSFromFile(cert, key)
		if err != nil {
			return nil, err
		}
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxPayloadSize),
			grpc.MaxCallSendMsgSize(maxPayloadSize),
		),
	}

	return grpc.Dial(server, opts...)
}
