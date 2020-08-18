package app

import (
	"github.com/BrobridgeOrg/gravity-adapter-stan/pkg/grpcbus"
)

type App interface {
	GetGRPCPool() grpcbus.GRPCPool
}
