package main

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"

	"github.com/1354092549/wsrpc"
)

var accountEventsToForward = []string{
	"on_receive_chat_message",
	"on_member_joined",
	"on_member_left",
	"process_group_invitation",
	"process_friend_request",
	"process_membership_request"}

type eventResult struct {
	Type int `json:"type"`
}

var accountAPIRPC = func() *wsrpc.WebsocketRPC {
	rpc := wsrpc.NewWebsocketRPC()
	for _, name := range accountEventsToForward {
		configAccountEventForward(rpc, name)
	}
	return rpc
}()

func configAccountEventForward(rpc *wsrpc.WebsocketRPC, name string) {
	forwarder := func(rpcConn *wsrpc.WebsocketRPCConn, arg json.RawMessage, reply *json.RawMessage) error {
		bot := rpcConn.Session["bot"].(string)
		params, err := AssembleForwardRequest("bot", bot, arg)
		if err != nil {
			return err
		}
		r := json.RawMessage([]byte(`{"type":0}`))
		ClientListMutex.RLock()
		apps := AllApps()
		ClientListMutex.RUnlock()
	AppLoop:
		for _, app := range apps {
			allConn := app.ConnList.All()
			for _, conn := range allConn {
				var rAppResult json.RawMessage
				appErr := conn.CallLowLevel(name, params, &rAppResult)
				if appErr != nil {
					continue
				}
				var appResult eventResult
				appErr = json.Unmarshal(rAppResult, &appResult)
				if appErr != nil {
					continue
				}
				// appResult.Type != 0 && appResult.Type != 1
				if appResult.Type&^1 != 0 {
					r = rAppResult
					break AppLoop
				}
			}
		}
		replyBytes, err := json.Marshal(r)
		if err != nil {
			return err
		}
		*reply = json.RawMessage(replyBytes)
		return nil
	}
	rpc.RegisterLowLevel(name, forwarder)
}

func AccountAPIWSHandler(query url.Values, adapter wsrpc.MessageAdapter) error {
	id := query.Get("id")
	token := query.Get("token")
	noRequest, _ := strconv.ParseBool(query.Get("norequest"))
	ClientListMutex.RLock()
	info, accountExists := Accounts[id]
	ClientListMutex.RUnlock()
	if !accountExists {
		return errors.New("id is not registered")
	}
	if info.Token != token {
		return errors.New("token is invalid")
	}
	rpcConn := accountAPIRPC.ConnectAdapter(adapter)
	rpcConn.Session["bot"] = id
	if noRequest {
		rpcConn.ServeConn()
	} else {
		elem := info.ConnList.Push(rpcConn)
		rpcConn.ServeConn()
		info.ConnList.Pop(elem)
	}
	return nil
}
