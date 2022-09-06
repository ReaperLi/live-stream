package api

import (
	"errors"
	"github.com/gorilla/websocket"
	"sync"
)

type Connection struct {
	wsConn      *websocket.Conn
	connManager *ConnManager
	inChan      chan []byte
	outChan     chan []byte
	closeChan   chan byte
	isClosed    bool
	mutex       sync.Mutex
}

func NeWConnection(wsConn *websocket.Conn, connManager *ConnManager) (conn *Connection, err error) {
	conn = &Connection{
		wsConn:      wsConn,
		connManager: connManager,
		inChan:      make(chan []byte, 1000),
		outChan:     make(chan []byte, 1000),
		closeChan:   make(chan byte, 1),
		isClosed:    false,
	}
	go conn.readLoop()
	go conn.writeLoop()
	return
}

func (conn *Connection) ReadMessage() (data []byte, err error) {
	select {
	case data = <-conn.inChan:
	case <-conn.closeChan:
		err = errors.New("connection is closed")
	}

	return
}

func (conn *Connection) WriteMessage(data []byte) (err error) {
	select {
	case conn.outChan <- data:
	case <-conn.closeChan:
		err = errors.New("connection is closed")
	}

	return
}

func (conn *Connection) Close() {
	//线程安全的
	conn.wsConn.Close()

	conn.mutex.Lock()
	if !conn.isClosed {
		close(conn.closeChan)
		//close(conn.inChan)
		//close(conn.outChan)
		conn.isClosed = true
		delete(conn.connManager.Conns, conn)
	}
	conn.mutex.Unlock()
}

//内部实现

func (conn *Connection) readLoop() {
	var (
		data []byte
		err  error
	)

	for {
		_, data, err = conn.wsConn.ReadMessage()
		if err != nil {
			goto ERR
		}

		select {
		case conn.inChan <- data:
		case <-conn.closeChan:
			//closeChan关闭的时候
			goto ERR
		}
	}

ERR:
	conn.Close()
}

func (conn *Connection) writeLoop() {
	var (
		data []byte
		err  error
	)
	for {
		select {
		case data = <-conn.outChan:
		case <-conn.closeChan:
			//closeChan关闭的时候
			goto ERR
		}
		err = conn.wsConn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			goto ERR
		}

	}
ERR:
	conn.Close()
}
