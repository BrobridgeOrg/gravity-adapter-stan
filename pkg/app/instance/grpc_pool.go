package instance

import (
	"time"

	"github.com/BrobridgeOrg/gravity-adapter-stan/pkg/grpcbus"
	"github.com/BrobridgeOrg/gravity-adapter-stan/pkg/grpcbus/pool"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func (a *AppInstance) initGRPCPool() error {

	host := viper.GetString("dsa.host")

	log.WithFields(log.Fields{
		"host": host,
	}).Info("Initializing gRPC pool")

	options := &pool.Options{
		InitCap:     8,
		MaxCap:      16,
		DialTimeout: time.Second * 20,
		IdleTimeout: time.Second * 60,
	}

	// Initialize connection pool
	p, err := pool.NewGRPCPool(host, options, grpc.WithInsecure())
	if err != nil {
		return err
	}

	if p == nil {
		return err
	}

	a.grpcPool = p
	return nil
}

func (a *AppInstance) GetGRPCPool() grpcbus.GRPCPool {
	return grpcbus.GRPCPool(a.grpcPool)
}
