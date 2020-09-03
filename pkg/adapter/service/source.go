package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	eventbus "github.com/BrobridgeOrg/gravity-adapter-stan/pkg/eventbus/service"
	dsa "github.com/BrobridgeOrg/gravity-api/service/dsa"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	log "github.com/sirupsen/logrus"
)

var defaultInfo = SourceInfo{
	DurableName:         "DefaultGravity",
	PingInterval:        10,
	MaxPingsOutstanding: 3,
	MaxReconnects:       -1,
}

type Packet struct {
	EventName string      `json:"event"`
	Payload   interface{} `json:"payload"`
}

type Source struct {
	adapter             *Adapter
	eventBus            *eventbus.EventBus
	name                string
	host                string
	port                int
	clusterID           string
	durableName         string
	channel             string
	pingInterval        int64
	maxPingsOutstanding int
	maxReconnects       int
}

func NewSource(adapter *Adapter, name string, sourceInfo *SourceInfo) *Source {

	// required cluster ID
	if len(sourceInfo.ClusterID) == 0 {
		log.WithFields(log.Fields{
			"source": name,
		}).Error("Required cluster ID")

		return nil
	}

	// required channel
	if len(sourceInfo.Channel) == 0 {
		log.WithFields(log.Fields{
			"source": name,
		}).Error("Required channel")

		return nil
	}

	info := sourceInfo

	// default settings
	if defaultInfo.DurableName != info.DurableName {
		info.DurableName = defaultInfo.DurableName
	}

	if defaultInfo.PingInterval != info.PingInterval {
		info.PingInterval = defaultInfo.PingInterval
	}

	if defaultInfo.MaxPingsOutstanding != info.MaxPingsOutstanding {
		info.MaxPingsOutstanding = defaultInfo.MaxPingsOutstanding
	}

	if defaultInfo.MaxReconnects != info.MaxReconnects {
		info.MaxReconnects = defaultInfo.MaxReconnects
	}

	return &Source{
		adapter:             adapter,
		name:                name,
		host:                info.Host,
		port:                info.Port,
		clusterID:           info.ClusterID,
		durableName:         info.DurableName,
		channel:             info.Channel,
		pingInterval:        info.PingInterval,
		maxPingsOutstanding: info.MaxPingsOutstanding,
		maxReconnects:       info.MaxReconnects,
	}
}

func (source *Source) InitSubscription() error {

	// Subscribe to channel
	stanConn := source.eventBus.GetSTANConnection()
	if len(source.durableName) == 0 {

		// Subscribe without durable name
		_, err := stanConn.Subscribe(source.channel, source.HandleMessage, stan.SetManualAckMode())
		if err != nil {
			return err
		}

		return nil
	}

	// Subscribe with durable name
	_, err := stanConn.Subscribe(source.channel, source.HandleMessage, stan.DurableName(source.durableName), stan.SetManualAckMode())
	if err != nil {
		return err
	}

	return nil
}

func (source *Source) Init() error {

	address := fmt.Sprintf("%s:%d", source.host, source.port)

	log.WithFields(log.Fields{
		"source":      source.name,
		"address":     address,
		"client_name": source.adapter.clientID + "-" + source.name,
		"cluster_id":  source.clusterID,
		"durableName": source.durableName,
		"channel":     source.channel,
	}).Info("Initializing source connector")

	options := eventbus.Options{
		ClusterID:           source.clusterID,
		ClientName:          source.adapter.clientID + "-" + source.name,
		PingInterval:        time.Duration(source.pingInterval),
		MaxPingsOutstanding: source.maxPingsOutstanding,
		MaxReconnects:       source.maxReconnects,
	}

	source.eventBus = eventbus.NewEventBus(
		address,
		eventbus.EventBusHandler{
			Reconnect: func(natsConn *nats.Conn) {
				err := source.InitSubscription()
				if err != nil {
					log.Error(err)
					return
				}

				log.Warn("re-connected to event server")
			},
			Disconnect: func(natsConn *nats.Conn) {
				log.Error("event server was disconnected")
			},
		},
		options,
	)

	err := source.eventBus.Connect()
	if err != nil {
		return err
	}

	return source.InitSubscription()
}

func (source *Source) HandleMessage(m *stan.Msg) {

	var packet Packet

	// Parse JSON
	err := json.Unmarshal(m.Data, &packet)
	if err != nil {
		m.Ack()
		return
	}

	log.WithFields(log.Fields{
		"event": packet.EventName,
		"seq":   m.Sequence,
	}).Info("Received event")

	// Convert payload to JSON string
	payload, err := json.Marshal(packet.Payload)
	if err != nil {
		m.Ack()
		return
	}

	request := &dsa.PublishRequest{
		EventName: packet.EventName,
		Payload:   string(payload),
	}

	// Getting connection from pool
	conn, err := source.adapter.app.GetGRPCPool().Get()
	if err != nil {
		log.Error("Failed to get connection: ", err)
		return
	}
	client := dsa.NewDataSourceAdapterClient(conn)

	// Preparing context
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Publish
	resp, err := client.Publish(ctx, request)
	if err != nil {
		log.Error("did not connect: ", err)
		return
	}

	if resp.Success == false {
		log.Error("Failed to push message to data source adapter")
		return
	}

	m.Ack()
}
