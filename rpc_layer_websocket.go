package main

import (
	"net/http"

	"github.com/1354092549/wsrpc"
	"github.com/gorilla/websocket"
)

func rpcCheckOrigin(r *http.Request) bool {
	return true
}

var wsRPCUpgrader = websocket.Upgrader{
	CheckOrigin:       rpcCheckOrigin,
	EnableCompression: true,
}

func (f RPCSessionHandler) WSRPCHandler(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	c, err := wsRPCUpgrader.Upgrade(writer, request, nil)
	if err != nil {
		return
	}
	defer c.Close()
	err = f(query, wsrpc.NewWebsocketMessageAdapter(c))
	if err != nil {
		c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, err.Error())) //nolint:errcheck
		return
	}
}
