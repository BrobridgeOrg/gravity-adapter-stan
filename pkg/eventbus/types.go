package eventbus

import (
	nats "github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
)

type EventBus interface {
	Connect() error
	Close()
	GetConnection() *nats.Conn
	GetSTANConnection() stan.Conn
}
