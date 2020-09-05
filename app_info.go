package main

type AppInfo struct {
	ID              string
	Token           string
	ManagerMetadata map[string]string
	ConnList        RPCConnList
}
