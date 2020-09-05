package main

import (
	"crypto/rand"
	"io"
	"sync"
	"unicode"

	"github.com/nicksnyder/basen"
)

func NewToken() string {
	var u [16]byte
	_, _ = io.ReadFull(rand.Reader, u[:])
	return basen.Base62Encoding.EncodeToString(u[:])
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
	ClientListMutex.RLock()
	r := make([]*AppInfo, len(Apps))
	i := 0
	for _, app := range Apps {
		r[i] = app
		i++
	}
	ClientListMutex.RUnlock()
	return r
}
