package main

import (
	"github.com/JimmeyLiu/pwx/pkg/transport"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	server := transport.NewServer(&TestHandler{})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		server.HandleWs(conn)
	})
	log.Println("Server started on :48080")
	_ = http.ListenAndServe(":48080", nil)
}

type TestHandler struct {
}

func (h *TestHandler) Path() string {
	return "test"
}

func (h *TestHandler) Handle(req transport.Request) transport.Response {
	return transport.Response{
		Status: 200,
		Body:   []byte("hello first request"),
	}
}

func (h *TestHandler) HandleStream(req transport.Request, callback func(resp transport.Response)) {
	callback(transport.Response{
		Status: 100,
		Body:   []byte("hello first request"),
	})
}
