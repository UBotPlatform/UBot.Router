package main

import (
	"container/list"
	"sync"

	"github.com/1354092549/wsrpc"
)

type RPCConnList struct {
	rw   sync.RWMutex
	conn list.List
}

func (l *RPCConnList) Push(conn *wsrpc.WebsocketRPCConn) *list.Element {
	l.rw.Lock()
	elem := l.conn.PushBack(conn)
	l.rw.Unlock()
	return elem
}

func (l *RPCConnList) Pop(elem *list.Element) {
	l.rw.Lock()
	l.conn.Remove(elem)
	l.rw.Unlock()
}

func (l *RPCConnList) All() []*wsrpc.WebsocketRPCConn {
	l.rw.RLock()
	r := make([]*wsrpc.WebsocketRPCConn, l.conn.Len())
	i := 0
	for e := l.conn.Front(); e != nil; e = e.Next() {
		r[i] = e.Value.(*wsrpc.WebsocketRPCConn)
		i++
	}
	l.rw.RUnlock()
	return r
}
