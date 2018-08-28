package mqpool

import (
	"errors"
)

var (
	ErrClosed = errors.New("pool is closed")
)

type Connection interface {
	Close() error
}

type Pool interface {
	Get() (*Conn, error)
	Close()
	Len() int
}
