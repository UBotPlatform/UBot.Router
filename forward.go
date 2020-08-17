package main

import (
	"encoding/json"

	"github.com/1354092549/wsrpc"
)

func DisassembleForwardRequest(idName string, arg json.RawMessage) (string, json.RawMessage, error) {
	var id string
	var err error
	var params json.RawMessage
	if wsrpc.IsJSONArray(arg) {
		var argInArray []json.RawMessage
		err = json.Unmarshal(arg, &argInArray)
		if err != nil {
			return "", nil, err
		}
		err = json.Unmarshal(argInArray[0], &id)
		if err != nil {
			return "", nil, err
		}
		paramsBytes, err := json.Marshal(argInArray[1:])
		if err != nil {
			return "", nil, err
		}
		params = json.RawMessage(paramsBytes)
	} else {
		var argInMap map[string]json.RawMessage
		err = json.Unmarshal(arg, &argInMap)
		if err != nil {
			return "", nil, err
		}
		rawID, ok := argInMap[idName]
		if !ok {
			return "", nil, wsrpc.RPCInvalidParamsError
		}
		err = json.Unmarshal(rawID, &id)
		if err != nil {
			return "", nil, err
		}
		delete(argInMap, idName)
		paramsBytes, err := json.Marshal(argInMap)
		if err != nil {
			return "", nil, err
		}
		params = json.RawMessage(paramsBytes)
	}
	return id, params, nil
}

func AssembleForwardRequest(idName string, id string, arg json.RawMessage) (json.RawMessage, error) {
	var err error
	var params json.RawMessage
	rawIDBytes, err := json.Marshal(id)
	if err != nil {
		return nil, err
	}
	rawID := json.RawMessage(rawIDBytes)
	if wsrpc.IsJSONArray(arg) {
		var originalArgInArray []json.RawMessage
		err = json.Unmarshal(arg, &originalArgInArray)
		if err != nil {
			return nil, err
		}
		newArgInArray := make([]json.RawMessage, len(originalArgInArray)+1)
		newArgInArray[0] = rawID
		copy(newArgInArray[1:], originalArgInArray)
		paramsBytes, err := json.Marshal(newArgInArray)
		if err != nil {
			return nil, err
		}
		params = json.RawMessage(paramsBytes)
	} else {
		var argInMap map[string]json.RawMessage
		err = json.Unmarshal(arg, &argInMap)
		if err != nil {
			return nil, err
		}
		argInMap[idName] = rawID
		paramsBytes, err := json.Marshal(argInMap)
		if err != nil {
			return nil, err
		}
		params = json.RawMessage(paramsBytes)
	}
	return params, nil
}
