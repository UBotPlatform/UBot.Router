package main

import (
	"crypto/rand"
	"io"
	"sync"
	"unicode"

	"github.com/nicksnyder/basen"
	orderedmap "github.com/wk8/go-ordered-map"
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
var Apps = orderedmap.New()
var Accounts = orderedmap.New()
var Addr string

func AllApps() []*AppInfo {
	ClientListMutex.RLock()
	r := make([]*AppInfo, Apps.Len())
	i := 0
	for pair := Apps.Oldest(); pair != nil; pair = pair.Next() {
		app := pair.Value.(*AppInfo)
		r[i] = app
		i++
	}
	ClientListMutex.RUnlock()
	return r
}
