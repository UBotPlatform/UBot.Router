package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
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

func (f RPCSessionHandler) HTTPPostRPCHandler(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	query.Set("norequest", "1")
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	httpAddr := newHTTPRequestAdapter(body, writer)
	err = f(query, httpAddr)
	if err != nil && !httpAddr.WroteOnce {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}
