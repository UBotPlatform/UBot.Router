package main

import (
	"errors"
	"net/url"

	"github.com/1354092549/wsrpc"
)

var managerRpc = func() *wsrpc.WebsocketRPC {
	rpc := wsrpc.NewWebsocketRPC()
	rpc.Register("register_app", registerApp, []string{"id"}, nil)
	rpc.Register("register_account", registerAccount, []string{"id"}, nil)
	return rpc
}()

func registerApp(id string) string {
	ClientListMutex.Lock()
	info, ok := Apps[id]
	if !ok {
		info = &AppInfo{ID: id, Token: NewToken()}
		Apps[id] = info
	}
	ClientListMutex.Unlock()
	return info.Token
}

func registerAccount(id string) string {
	ClientListMutex.Lock()
	info, ok := Accounts[id]
	if !ok {
		info = &AccountInfo{ID: id, Token: NewToken()}
		Accounts[id] = info
	}
	ClientListMutex.Unlock()
	return info.Token
}

func ManagerAPISessionFactory(query url.Values, adapter wsrpc.MessageAdapter) error {
	if EnableAuthentication {
		token := query.Get("token")
		if token != ManagerToken {
			return errors.New("token is invalid")
		}
	}
	rpcConn := managerRpc.ConnectAdapter(adapter)
	rpcConn.ServeConn()
	return nil
}
