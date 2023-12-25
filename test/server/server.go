package main

import (
	"github.com/JimmeyLiu/pwx/pkg/transport"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	server := transport.NewServer()
	server.RegisterHandler(&TestHandler{})
	_ = server.Listen("0.0.0.0:9876")
	i := make(chan os.Signal)
	signal.Notify(i, syscall.SIGKILL)
	for {
		select {
		case <-i:
			return
		}
	}
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
