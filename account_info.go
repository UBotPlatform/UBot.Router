package main

type AccountInfo struct {
	ID              string
	Token           string
	ManagerMetadata map[string]string
	ConnList        RPCConnList
}
