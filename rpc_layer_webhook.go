package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/1354092549/wsrpc"
)

type WebhookInfo struct {
	Token   string `json:"token,omitempty"`
	Address string `json:"address,omitempty"`
}

type WebhookAdapter struct {
	Token   string
	Address string
	msgChan chan []byte
}

func NewWebhookAdapter(token string, address string) *WebhookAdapter {
	return &WebhookAdapter{
		Token:   token,
		Address: address,
		msgChan: make(chan []byte),
	}
}

func (a *WebhookAdapter) ReadMessage() ([]byte, error) {
	r := <-a.msgChan
	if r == nil {
		return nil, io.EOF
	}
	return r, nil
}

var webhookUA = func() string {
	var builder strings.Builder
	builder.WriteString("UBot.Router (")
	builder.WriteString(Upperfirst(runtime.GOOS))
	builder.WriteString("; ")
	builder.WriteString(runtime.GOARCH)
	builder.WriteString(")")
	builder.WriteString(" Golang (")
	builder.WriteString(runtime.Version())
	builder.WriteString(")")
	return builder.String()
}()

func (a *WebhookAdapter) WriteMessage(data []byte) error {
	req, err := http.NewRequest("POST", a.Address, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", webhookUA)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to invoke webhook, address: %s, detail: %v\n", a.Address, err)
		fmt.Fprintln(os.Stderr, "Note that this can severely slow down the efficiency of the event scheduler.")
		return err
	}
	defer resp.Body.Close()
	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if len(msg) != 0 {
		a.msgChan <- msg
	}
	return nil
}

func (a *WebhookAdapter) Close() error {
	a.msgChan <- nil
	return nil
}

var webhooks sync.Map

func createWebhook(controllerConn *wsrpc.WebsocketRPCConn, address string) (*WebhookInfo, error) {
	parsedAddr, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("invalid webhook address: %w", err)
	}
	if !strings.EqualFold(parsedAddr.Scheme, "http") && !strings.EqualFold(parsedAddr.Scheme, "https") {
		return nil, errors.New("invalid webhook address: only HTTP and HTTPS addresses are supported")
	}
	handler := controllerConn.Session["_handler"].(RPCSessionHandler)
	query := controllerConn.Session["_query"].(url.Values)
	adapter := NewWebhookAdapter(NewToken(), address)
	webhooks.Store(adapter.Token, adapter)
	go func() {
		_ = handler(query, adapter)
		webhooks.Delete(adapter.Token)
	}()
	return &WebhookInfo{Token: adapter.Token, Address: adapter.Address}, nil
}

func deleteWebhook(controllerConn *wsrpc.WebsocketRPCConn, token string) error {
	rAdapter, exists := webhooks.Load(token)
	if !exists {
		return errors.New("webhook is not found")
	}
	adapter := rAdapter.(*WebhookAdapter)
	return adapter.Close()
}

var webhookControllerRPC = func() *wsrpc.WebsocketRPC {
	r := wsrpc.NewWebsocketRPC()
	r.Register("create_webhook", createWebhook, []string{"address"}, nil)
	r.Register("delete_webhook", deleteWebhook, []string{"token"}, nil)
	return r
}()

func WebhookControllerSessionFactory(handler RPCSessionHandler) RPCSessionHandler {
	return func(query url.Values, adapter wsrpc.MessageAdapter) error {
		internalQuery := make(url.Values)
		for name, values := range query {
			if name == "webhook" || name == "norequest" {
				continue
			}
			internalQuery[name] = values
		}
		controllerConn := webhookControllerRPC.ConnectAdapter(adapter)
		controllerConn.Session["_handler"] = handler
		controllerConn.Session["_query"] = internalQuery
		controllerConn.ServeConn()
		return nil
	}
}
