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
	rInfo, ok := Apps.Get(id)
	var info *AppInfo
	if ok {
		info = rInfo.(*AppInfo)
	} else {
		info = &AppInfo{ID: id, Token: NewToken(), ManagerMetadata: getManagerMetadata(conn)}
		Apps.Set(id, info)
	}
	ClientListMutex.Unlock()
	return info.Token
}

func registerAccount(conn *wsrpc.WebsocketRPCConn, id string) string {
	ClientListMutex.Lock()
	rInfo, ok := Accounts.Get(id)
	var info *AccountInfo
	if ok {
		info = rInfo.(*AccountInfo)
	} else {
		info = &AccountInfo{ID: id, Token: NewToken(), ManagerMetadata: getManagerMetadata(conn)}
		Accounts.Set(id, info)
	}
	ClientListMutex.Unlock()
	return info.Token
}

func getAppList(conn *wsrpc.WebsocketRPCConn) []appInfoResponse {
	var r []appInfoResponse
	ClientListMutex.RLock()
	for pair := Apps.Oldest(); pair != nil; pair = pair.Next() {
		item := pair.Value.(*AppInfo)
		r = append(r, appInfoResponse{ID: item.ID, Token: item.Token, ManagerMetadata: item.ManagerMetadata})
	}
	ClientListMutex.RUnlock()
	return r
}

func getAccountList(conn *wsrpc.WebsocketRPCConn) []accountInfoResponse {
	var r []accountInfoResponse
	ClientListMutex.RLock()
	for pair := Accounts.Oldest(); pair != nil; pair = pair.Next() {
		item := pair.Value.(*AccountInfo)
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
