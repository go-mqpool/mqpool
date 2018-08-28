package mqpool

import (
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/streadway/amqp"
)

var (
	ErrConnectionString = errors.New("incorrect connection string")
	ErrParameterLength  = errors.New("incorrect parameter length")
	ErrConfig           = errors.New("incorrect config type")
)

type Factory func() (*amqp.Connection, error)

func NewConnPool(initialCap, maxCap int, args ...interface{}) (Pool, error) {

	var (
		factory Factory
		url     string
		tlsCfg  *tls.Config
		amqpCfg amqp.Config
		c       bool
	)
	switch len(args) {
	case 2:
		url, c = args[0].(string)
		if !c {
			return nil, ErrConnectionString
		}
		if args[1] == nil {
			return nil, ErrConfig
		}
		tlsCfg, c = args[1].(*tls.Config)
		if c {
			factory = func() (*amqp.Connection, error) { return amqp.DialTLS(url, tlsCfg) }
		}
		amqpCfg, c = args[1].(amqp.Config)
		if c {
			factory = func() (*amqp.Connection, error) { return amqp.DialConfig(url, amqpCfg) }
		}
		return nil, ErrConfig
	case 1:
		url, c = args[0].(string)
		if !c {
			return nil, ErrConnectionString
		}
		factory = func() (*amqp.Connection, error) { return amqp.Dial(url) }
	default:
		return nil, ErrParameterLength
	}
	return newConnPool(initialCap, maxCap, factory)
}

func newConnPool(initialCap, maxCap int, factory Factory) (Pool, error) {
	if initialCap < 0 || maxCap <= 0 || initialCap > maxCap {
		return nil, errors.New("invalid capacity settings")
	}

	c := &connectionPool{
		conns:   make(chan *amqp.Connection, maxCap),
		factory: factory,
	}

	for i := 0; i < initialCap; i++ {
		conn, err := factory()
		if err != nil {
			c.Close()
			return nil, fmt.Errorf("factory is not able to fill the pool: %s", err)
		}
		c.conns <- conn
	}

	return c, nil
}
