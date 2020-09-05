package main

import (
	"errors"
	"net/url"
	"strings"

	"github.com/1354092549/wsrpc"
)

type appInfoResponse struct {
	ID              string            `json:"id"`
	Token           string            `json:"token"`
	ManagerMetadata map[string]string `json:"manager_metadata"`
}

type accountInfoResponse struct {
	ID              string            `json:"id"`
	Token           string            `json:"token"`
	ManagerMetadata map[string]string `json:"manager_metadata"`
}

var managerRpc = func() *wsrpc.WebsocketRPC {
	rpc := wsrpc.NewWebsocketRPC()
	rpc.Register("register_app", registerApp, []string{"id"}, nil)
	rpc.Register("register_account", registerAccount, []string{"id"}, nil)
	rpc.Register("get_app_list", getAppList, nil, nil)
	rpc.Register("get_account_list", getAccountList, nil, nil)
	return rpc
}()

func getManagerMetadata(conn *wsrpc.WebsocketRPCConn) map[string]string {
	return conn.Session["ManagerMetadata"].(map[string]string)
}

func registerApp(conn *wsrpc.WebsocketRPCConn, id string) string {
	ClientListMutex.Lock()
	info, ok := Apps[id]
	if !ok {
		info = &AppInfo{ID: id, Token: NewToken(), ManagerMetadata: getManagerMetadata(conn)}
		Apps[id] = info
	}
	ClientListMutex.Unlock()
	return info.Token
}

func registerAccount(conn *wsrpc.WebsocketRPCConn, id string) string {
	ClientListMutex.Lock()
	info, ok := Accounts[id]
	if !ok {
		info = &AccountInfo{ID: id, Token: NewToken(), ManagerMetadata: getManagerMetadata(conn)}
		Accounts[id] = info
	}
	ClientListMutex.Unlock()
	return info.Token
}

func getAppList(conn *wsrpc.WebsocketRPCConn) []appInfoResponse {
	var r []appInfoResponse
	ClientListMutex.RLock()
	for _, item := range Apps {
		r = append(r, appInfoResponse{ID: item.ID, Token: item.Token, ManagerMetadata: item.ManagerMetadata})
	}
	ClientListMutex.RUnlock()
	return r
}

func getAccountList(conn *wsrpc.WebsocketRPCConn) []accountInfoResponse {
	var r []accountInfoResponse
	ClientListMutex.RLock()
	for _, item := range Accounts {
		r = append(r, accountInfoResponse{ID: item.ID, Token: item.Token, ManagerMetadata: item.ManagerMetadata})
	}
	ClientListMutex.RUnlock()
	return r
}

func ManagerAPISessionFactory(query url.Values, adapter wsrpc.MessageAdapter) error {
	if EnableAuthentication {
		token := query.Get("token")
		if token != ManagerToken {
			return errors.New("token is invalid")
		}
	}
	rpcConn := managerRpc.ConnectAdapter(adapter)
	metadata := make(map[string]string)
	for name, values := range query {
		if strings.HasPrefix(name, "x-") || strings.HasPrefix(name, "X-") {
			if len(values) > 0 {
				metadata[name] = values[0]
			} else {
				metadata[name] = ""
			}
		}
	}
	rpcConn.Session["ManagerMetadata"] = metadata
	rpcConn.ServeConn()
	return nil
}
