package main

import (
	"encoding/base64"
	"sync"
	"unicode"

	uuid "github.com/satori/go.uuid"
)

func NewToken() string {
	return base64.RawURLEncoding.EncodeToString(uuid.NewV4().Bytes())
}

func Upperfirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

var ManagerToken = NewToken()
var EnableAuthentication bool = false
var ClientListMutex sync.RWMutex
var Apps = make(map[string]*AppInfo)
var Accounts = make(map[string]*AccountInfo)
var Addr string

func AllApps() []*AppInfo {
	r := make([]*AppInfo, len(Apps))
	i := 0
	for _, app := range Apps {
		r[i] = app
		i++
	}
	return r
}
