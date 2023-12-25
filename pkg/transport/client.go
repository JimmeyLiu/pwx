package transport

import (
	"github.com/gorilla/websocket"
	"io"
	"log/slog"
	"net"
	"time"
)

type PwxClient struct {
	network         string
	addr            string
	rw              io.ReadWriter
	transport       pwxTransport
	lastConnectTime int64
}

func NewClient(addr string, handlers ...RequestHandler) *PwxClient {
	return newClient("tcp4", addr, handlers)
}

func NewUnixClient(file string, handlers ...RequestHandler) *PwxClient {
	return newClient("unix", file, handlers)
}

func NewWsClient(ws string, handlers ...RequestHandler) *PwxClient {
	return newClient("ws", ws, handlers)
}

func newClient(network, addr string, handlers []RequestHandler) *PwxClient {
	hs := make(map[string]RequestHandler)
	for _, h := range handlers {
		hs[h.Path()] = h
	}
	cli := &PwxClient{
		network:   network,
		addr:      addr,
		transport: newTransport(hs),
	}
	go cli.connect()
	return cli
}

func (c *PwxClient) Request(request Request, seconds int) Response {
	return c.transport.doRequest(request, seconds)
}

func (c *PwxClient) RequestStream(request Request, seconds int, callback func(resp Response)) {
	c.transport.doRequestStream(request, seconds, callback)
}

func (c *PwxClient) connect() {
	var rw io.ReadWriter
	if c.network == "ws" {
		conn, _, err := websocket.DefaultDialer.Dial(c.addr, nil)
		if err != nil {
			slog.Error("connect ws error: " + err.Error())
			return
		}
		rw = newWsReadWriter(conn)
	} else {
		conn, err := net.Dial(c.network, c.addr)
		if err != nil {
			slog.Error("connect error: " + err.Error())
			go c.reconnect()
			return
		}
		rw = conn
	}
	c.lastConnectTime = time.Now().UnixMilli()
	_ = c.transport.readWrite(rw)
}

func (c *PwxClient) reconnect() {
	if time.Now().UnixMilli()-c.lastConnectTime < 10 {
		time.AfterFunc(time.Second*10, func() {
			c.connect()
		})
	} else {
		c.connect()
	}
}
