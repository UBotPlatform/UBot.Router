package main

import (
	"flag"
	"fmt"
	"net/http"
)

var managerUser string
var managerPassword string

func main() {
	fmt.Println("UBot Router is running")
	flag.StringVar(&managerUser, "user", "", "")
	flag.StringVar(&managerPassword, "password", "", "")
	flag.StringVar(&Addr, "addr", "localhost:5000", "")
	flag.Parse()
	if managerUser == "" && managerPassword == "" {
		fmt.Println("No password configured")
	} else {
		EnableAuthentication = true
		fmt.Println("Password configured")
	}
	fmt.Println("Address: " + Addr)
	router := http.NewServeMux()
	router.HandleFunc("/", WelcomeHandler)
	router.HandleFunc("/api/manager/get_token", GetManagerTokenHandler)
	router.HandleFunc("/api/manager", RPCHandler(ManagerAPISessionFactory))
	router.HandleFunc("/api/account", RPCHandler(AccountAPIWSHandler))
	router.HandleFunc("/api/app", RPCHandler(AppApiSessionFactory))
	err := http.ListenAndServe(Addr, CORSMiddleware(router))
	if err != nil {
		fmt.Println("Cannot listen or serve: " + err.Error())
	}
}

func GetManagerTokenHandler(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	if len(managerUser) != 0 || len(managerPassword) != 0 {
		user := request.Form.Get("user")
		password := request.Form.Get("password")
		if (managerUser != "" && managerUser != user) || (managerPassword != "" && managerPassword != password) {
			http.Error(writer, "User or password is incorrect", http.StatusForbidden)
			return
		}
	}
	writer.Header().Add("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte(ManagerToken))
}

func WelcomeHandler(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" {
		http.NotFound(writer, request)
		return
	}
	writer.WriteHeader(http.StatusOK)
	writer.Header().Add("Content-Type", "text/plain")
	_, _ = writer.Write([]byte("Welcome to use UBot.Router"))
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, AccessToken, X-CSRF-Token, Authorization, Token")
		writer.Header().Set("Access-Control-Allow-Credentials", "true")
		writer.Header().Set("Access-Control-Allow-Methods", "HEAD, GET, POST, PUT, PATCH, DELETE")
		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
