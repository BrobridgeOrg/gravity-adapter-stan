package adapter

import (
	"fmt"
	"os"
	"strings"

	"github.com/BrobridgeOrg/gravity-adapter-stan/pkg/app"
	log "github.com/sirupsen/logrus"
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

	// Using hostname (pod name) by default
	host, err := os.Hostname()
	if err != nil {
		log.Error(err)
		return nil
	}

	host = strings.ReplaceAll(host, ".", "_")

	adapter.clientID = fmt.Sprintf("gravity-%s", host)

	err = adapter.sm.Initialize()
	if err != nil {
		log.Error(err)
		return nil
	}

	return nil
}
