package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	EnableCompression: false,
}

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
	router := mux.NewRouter()
	router.HandleFunc("/", WelcomeHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/manager/get_token", GetManagerTokenHandler).Methods(http.MethodPost)

	router.HandleFunc("/api/manager", WSRPCHandler(ManagerAPISessionFactory)).Methods(http.MethodGet)
	router.HandleFunc("/api/account", WSRPCHandler(AccountAPIWSHandler)).Methods(http.MethodGet)
	router.HandleFunc("/api/app", WSRPCHandler(AppApiSessionFactory)).Methods(http.MethodGet)

	router.HandleFunc("/api/manager", HTTPPostRPCHandler(ManagerAPISessionFactory)).Methods(http.MethodPost)
	router.HandleFunc("/api/account", HTTPPostRPCHandler(AccountAPIWSHandler)).Methods(http.MethodPost)
	router.HandleFunc("/api/app", HTTPPostRPCHandler(AppApiSessionFactory)).Methods(http.MethodPost)

	router.Use(CORSMiddleware)

	http.Handle("/", router)
	err := http.ListenAndServe(Addr, nil)
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
	writer.WriteHeader(http.StatusOK)
	writer.Header().Add("Content-Type", "text/plain")
	_, _ = writer.Write([]byte("Welcome to use UBot.Router"))
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, AccessToken, X-CSRF-Token, Authorization, Token")
		writer.Header().Set("Access-Control-Allow-Credentials", "true")
		writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
