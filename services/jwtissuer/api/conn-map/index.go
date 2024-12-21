package connmap

import (
	"sync"

	"github.com/gorilla/websocket"
)

type ConnMap struct {
	conns map[string]*websocket.Conn
	mutex *sync.RWMutex
}

func New() *ConnMap {
	return &ConnMap{
		conns: make(map[string]*websocket.Conn),
		mutex: new(sync.RWMutex),
	}
}

func (c *ConnMap) Get(id string) (*websocket.Conn, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	conn, ok := c.conns[id]
	return conn, ok
}

func (c *ConnMap) Set(id string, conn *websocket.Conn) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.conns[id] = conn
}

func (c *ConnMap) Del(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	conn, ok := c.conns[id]
	if ok {
		conn.Close()
		delete(c.conns, id)
	}
}
