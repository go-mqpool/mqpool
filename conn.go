package mqpool

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/streadway/amqp"
)

type Conn struct {
	*amqp.Connection
	mu       sync.RWMutex
	c        *connectionPool
	unusable bool
}

func (p *Conn) Close() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.unusable {
		if p.Connection != nil {
			defer func() {
				v := reflect.ValueOf(p.Connection)
				x := v.Elem().FieldByName("closed")

				fmt.Println("isclose:", x)
			}()
			return p.Connection.Close()
		}
		return nil
	}
	return p.c.put(p.Connection)
}

func (p *Conn) MarkUnusable() {
	p.mu.Lock()
	p.unusable = true
	p.mu.Unlock()
}

func (c *connectionPool) wrapConn(conn *amqp.Connection) *Conn {
	p := &Conn{c: c}
	p.Connection = conn
	return p
}
