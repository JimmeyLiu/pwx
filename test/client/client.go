package main

import (
	"github.com/JimmeyLiu/pwx/pkg/transport"
	"log"
	"time"
)

func main() {

	//cli := transport.NewClient("0.0.0.0:9876")
	//time.Sleep(3 * time.Second)
	//resp := cli.Request(transport.Request{
	//	Path: "test",
	//}, 10)
	//log.Println("resp ", resp.Status)
	//cli.RequestStream(transport.Request{
	//	Path: "test",
	//}, 10, func(resp transport.Response) {
	//	log.Println(fmt.Sprintf("resp %d %s", resp.Status, string(resp.Body)))
	//})

	cli := transport.NewWsClient("ws://127.0.0.1:48080/ws")
	time.Sleep(1 * time.Second)
	resp := cli.Request(transport.Request{
		Path: "test",
	}, 10)
	log.Println("ws resp ", resp.Status)
}
