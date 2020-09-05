package main

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/1354092549/wsrpc"
)

type RPCSessionHandler func(query url.Values, adapter wsrpc.MessageAdapter) error

func RPCHandler(factory RPCSessionHandler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		query := request.URL.Query()
		webhook, _ := strconv.ParseBool(query.Get("webhook"))
		actualFactory := factory
		if webhook {
			actualFactory = WebhookControllerSessionFactory(factory)
		}
		switch request.Method {
		case http.MethodGet:
			actualFactory.WSRPCHandler(writer, request)
		case http.MethodPost:
			actualFactory.HTTPPostRPCHandler(writer, request)
		}
	}
}
