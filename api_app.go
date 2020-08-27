package main

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"

	"github.com/1354092549/wsrpc"
)

var appAPIMethodsToForward = []string{
	"send_chat_message",
	"get_user_name",
	"get_group_name",
	"remove_member",
	"shutup_member",
	"shutup_all_member",
	"get_member_name",
	"get_user_avatar",
	"get_self_id",
	"get_platform_id",
	"get_group_list",
	"get_member_list"}

var appAPIRPC = func() *wsrpc.WebsocketRPC {
	rpc := wsrpc.NewWebsocketRPC()
	for _, name := range appAPIMethodsToForward {
		configAppAPIForward(rpc, name)
	}
	rpc.Register("get_bot_list", getBotList, nil, nil)
	return rpc
}()

func configAppAPIForward(rpc *wsrpc.WebsocketRPC, name string) {
	forwarder := func(rpcConn *wsrpc.WebsocketRPCConn, arg json.RawMessage, reply *json.RawMessage) error {
		bot, params, err := DisassembleForwardRequest("bot", arg)
		if err != nil {
			return err
		}
		ClientListMutex.RLock()
		info, accountExists := Accounts[bot]
		ClientListMutex.RUnlock()
		if !accountExists {
			return wsrpc.RPCInvalidParamsError
		}
		allConn := info.ConnList.All()
		for _, conn := range allConn {
			err = conn.CallLowLevel(name, params, reply)
			if err == nil {
				return nil
			}
		}
		return err
	}
	rpc.RegisterLowLevel(name, forwarder)
}

func getBotList() ([]string, error) {
	var r []string
	ClientListMutex.RLock()
	for _, account := range Accounts {
		r = append(r, account.ID)
	}
	ClientListMutex.RUnlock()
	return r, nil
}

func AppApiSessionFactory(query url.Values, adapter wsrpc.MessageAdapter) error {
	id := query.Get("id")
	token := query.Get("token")
	noRequest, _ := strconv.ParseBool(query.Get("norequest"))
	ClientListMutex.RLock()
	info, appExists := Apps[id]
	ClientListMutex.RUnlock()
	if !appExists {
		return errors.New("id is not registered")
	}
	if info.Token != token {
		return errors.New("token is invalid")
	}
	rpcConn := appAPIRPC.ConnectAdapter(adapter)
	if noRequest {
		rpcConn.ServeConn()
	} else {
		elem := info.ConnList.Push(rpcConn)
		rpcConn.ServeConn()
		info.ConnList.Pop(elem)
	}
	return nil
}
