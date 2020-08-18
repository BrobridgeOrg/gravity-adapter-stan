package adapter

import (
	"strconv"

	"github.com/BrobridgeOrg/gravity-adapter-stan/pkg/app"
	log "github.com/sirupsen/logrus"
	"github.com/sony/sonyflake"
)

type Adapter struct {
	app      app.App
	sm       *SourceManager
	clientID string
}

func NewAdapter(a app.App) *Adapter {
	adapter := &Adapter{
		app: a,
	}

	adapter.sm = NewSourceManager(adapter)

	return adapter
}

func (adapter *Adapter) Init() error {

	// Genereate a unique ID for instance
	flake := sonyflake.NewSonyflake(sonyflake.Settings{})
	id, err := flake.NextID()
	if err != nil {
		return nil
	}

	adapter.clientID = strconv.FormatUint(id, 16)

	err = adapter.sm.Initialize()
	if err != nil {
		log.Error(err)
		return nil
	}

	return nil
}
