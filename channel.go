package mqpool

import (
	"errors"
	"sync"

	"github.com/streadway/amqp"
)

type connectionPool struct {
	mu      sync.RWMutex
	conns   chan *amqp.Connection
	factory Factory
}

func (c *connectionPool) getConnsAndFactory() (chan *amqp.Connection, Factory) {
	c.mu.RLock()
	conns := c.conns
	factory := c.factory
	c.mu.RUnlock()
	return conns, factory
}

func (c *connectionPool) Get() (*Conn, error) {
	conns, factory := c.getConnsAndFactory()
	if conns == nil {
		return nil, ErrClosed
	}

	select {
	case conn := <-conns:
		if conn == nil {
			return nil, ErrClosed
		}

		return c.wrapConn(conn), nil
	default:
		conn, err := factory()
		if err != nil {
			return nil, err
		}

		return c.wrapConn(conn), nil
	}
}

func (c *connectionPool) put(conn *amqp.Connection) error {
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conns == nil {
		return conn.Close()
	}

	select {
	case c.conns <- conn:
		return nil
	default:
		return conn.Close()
	}
}

func (c *connectionPool) Close() {
	c.mu.Lock()
	conns := c.conns
	c.conns = nil
	c.factory = nil
	c.mu.Unlock()

	if conns == nil {
		return
	}

	close(conns)
	for conn := range conns {
		conn.Close()
	}
}

func (c *connectionPool) Len() int {
	conns, _ := c.getConnsAndFactory()
	return len(conns)
}
