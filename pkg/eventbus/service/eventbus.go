package eventbus

import (
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	log "github.com/sirupsen/logrus"
)

type Options struct {
	ClusterID           string
	ClientName          string
	PingInterval        time.Duration
	MaxPingsOutstanding int
	MaxReconnects       int
}

type EventBusHandler struct {
	Reconnect  func(natsConn *nats.Conn)
	Disconnect func(natsConn *nats.Conn)
}

type EventBus struct {
	connection *nats.Conn
	stanConn   stan.Conn
	host       string
	handler    *EventBusHandler
	options    *Options
}

func NewEventBus(host string, handler EventBusHandler, options Options) *EventBus {
	return &EventBus{
		connection: nil,
		host:       host,
		handler:    &handler,
		options:    &options,
	}
}

func (eb *EventBus) Connect() error {

	log.WithFields(log.Fields{
		"host":                eb.host,
		"PingInterval":        eb.options.PingInterval * time.Second,
		"MaxPingsOutnatsding": eb.options.MaxPingsOutstanding,
		"MaxReconnects":       eb.options.MaxReconnects,
	}).Info("Connecting to NATS server")

	nc, err := nats.Connect(eb.host,
		nats.PingInterval(eb.options.PingInterval*time.Second),
		nats.MaxPingsOutstanding(eb.options.MaxPingsOutstanding),
		nats.MaxReconnects(eb.options.MaxReconnects),
		nats.ReconnectHandler(eb.ReconnectHandler),
		nats.DisconnectHandler(eb.handler.Disconnect),
	)
	if err != nil {
		return err
	}

	eb.connection = nc

	// Connect to NATS Streaming
	err = eb.ConnectToSTAN()
	if err != nil {
		return err
	}

	return nil
}

func (eb *EventBus) ConnectToSTAN() error {

	log.WithFields(log.Fields{
		"clusterID":  eb.options.ClusterID,
		"clientName": eb.options.ClientName,
	}).Info("Connecting to NATS Streaming")

	// Connect to streaming server
	sc, err := stan.Connect(
		eb.options.ClusterID,
		eb.options.ClientName,
		stan.NatsConn(eb.connection),
	)
	if err != nil {
		return err
	}

	eb.stanConn = sc

	return nil
}

func (eb *EventBus) Close() {
	eb.connection.Close()
}
func (eb *EventBus) ReconnectHandler(natsConn *nats.Conn) {

	// Reconnect to NATS Streaming
	err := eb.ConnectToSTAN()
	if err != nil {
		log.Error(err)
		return
	}

	eb.handler.Reconnect(natsConn)
}

func (eb *EventBus) GetConnection() *nats.Conn {
	return eb.connection
}

func (eb *EventBus) GetSTANConnection() stan.Conn {
	return eb.stanConn
}
