package api

import (
	"sync"
)

type ConnManager struct {
	Conns    map[*Connection]bool
	ConnLock sync.RWMutex
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		Conns: make(map[*Connection]bool),
	}
}

func (connManager *ConnManager) GetConns() map[*Connection]bool {
	return connManager.Conns
}

func (connManager *ConnManager) Register(conn *Connection) {
	connManager.ConnLock.Lock()
	defer connManager.ConnLock.Unlock()

	if connManager.Conns[conn] == false {
		connManager.Conns[conn] = true
	}
}

func (connManager *ConnManager) Unregister(conn *Connection) {
	conn.Close()
}
