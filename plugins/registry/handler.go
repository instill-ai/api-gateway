package main

import (
	"fmt"

	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
)

type registryHandler struct {
	mgmtPublicClient  mgmtPB.MgmtPublicServiceClient
	mgmtPrivateClient mgmtPB.MgmtPrivateServiceClient

	registryAddr string
}

func newRegistryHandler(config map[string]any) (*registryHandler, error) {
	var mgmtPublicAddr, mgmtPrivateAddr string
	var ok bool
	var rh registryHandler

	if rh.registryAddr, ok = config["hostport"].(string); !ok {
		return nil, fmt.Errorf("invalid registry address")
	}
	if mgmtPublicAddr, ok = config["mgmt_public_hostport"].(string); !ok {
		return nil, fmt.Errorf("invalid mgmt public address")
	}
	if mgmtPrivateAddr, ok = config["mgmt_private_hostport"].(string); !ok {
		return nil, fmt.Errorf("invalid mgmt private address")
	}

	mgmtPublicConn, err := newGRPCConn(mgmtPublicAddr, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to connect with mgmt-backend: %w", err)
	}
	mgmtPrivateConn, err := newGRPCConn(mgmtPrivateAddr, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to connect with mgmt-backend: %w", err)
	}

	rh.mgmtPublicClient = mgmtPB.NewMgmtPublicServiceClient(mgmtPublicConn)
	rh.mgmtPrivateClient = mgmtPB.NewMgmtPrivateServiceClient(mgmtPrivateConn)

	return &rh, nil
}
