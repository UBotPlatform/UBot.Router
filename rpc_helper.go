package main

import (
	"net/http"
	"net/url"

	"github.com/1354092549/wsrpc"
	"github.com/gorilla/websocket"
)

type RPCSessionHandler func(query url.Values, adapter wsrpc.MessageAdapter) error

func WSRPCHandler(handler RPCSessionHandler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		query := request.URL.Query()
		c, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		defer c.Close()
		err = handler(query, wsrpc.NewWebsocketMessageAdapter(c))
		if err != nil {
			c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, err.Error())) //nolint:errcheck
			return
		}
	}
}
