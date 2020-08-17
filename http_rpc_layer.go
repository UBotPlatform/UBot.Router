package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/1354092549/wsrpc"
)

type onlyIdMessage struct {
	ID json.RawMessage `json:"id"`
}

func isRPCNotification(data []byte) bool {
	var test onlyIdMessage
	err := json.Unmarshal(data, &test)
	return err == nil && test.ID == nil
}

type HTTPRequestAdapter struct {
	body      []byte
	writer    http.ResponseWriter
	waitChan  chan int
	WroteOnce bool
}

func newHTTPRequestAdapter(body []byte, writer http.ResponseWriter) *HTTPRequestAdapter {
	waitChan := make(chan int, 1)
	r := &HTTPRequestAdapter{body: body, writer: writer, waitChan: waitChan}
	if isRPCNotification(body) {
		r.waitChan <- 0
		r.WroteOnce = true
	}
	return r
}

func (a *HTTPRequestAdapter) ReadMessage() ([]byte, error) {
	if a.body == nil {
		<-a.waitChan
		return nil, io.EOF
	}
	r := a.body
	a.body = nil
	return r, nil
}

func (a *HTTPRequestAdapter) WriteMessage(data []byte) error {
	if a.WroteOnce {
		return errors.New("cannot write message twice since this is a once message adapter")
	}
	a.WroteOnce = true
	a.writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	a.writer.Header().Set("Content-Length", strconv.Itoa(len(data)))
	a.writer.WriteHeader(http.StatusOK)
	_, err := a.writer.Write(data)
	a.waitChan <- 0
	return err
}

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
		return nil, errors.New("invalid webhook address: " + err.Error())
	}
	if parsedAddr.Scheme != "http" && parsedAddr.Scheme != "https" {
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

func HTTPPostRPCHandler(handler RPCSessionHandler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		query := request.URL.Query()
		webhook, _ := strconv.ParseBool(query.Get("webhook"))
		if webhook {
			query.Del("webhook")
			query.Del("norequest")
		} else {
			query.Set("norequest", "1")
		}
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		httpAddr := newHTTPRequestAdapter(body, writer)
		if webhook {
			controllerConn := webhookControllerRPC.ConnectAdapter(httpAddr)
			controllerConn.Session["_handler"] = handler
			controllerConn.Session["_query"] = query
			controllerConn.ServeConn()
		} else {
			err = handler(query, httpAddr)
		}
		if err != nil && !httpAddr.WroteOnce {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
