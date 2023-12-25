package transport

import (
	"github.com/gorilla/websocket"
	"net"
)

type PwxServer struct {
	handlers map[string]RequestHandler
}

func NewServer(handlers ...RequestHandler) *PwxServer {
	hs := make(map[string]RequestHandler)
	for _, h := range handlers {
		hs[h.Path()] = h
	}
	return &PwxServer{
		handlers: hs,
	}
}

func (p *PwxServer) RegisterHandler(h RequestHandler) {
	p.handlers[h.Path()] = h
}

func (p *PwxServer) Listen(addr string) error {
	l, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}
	go p.accept(l)
	return nil
}

func (p *PwxServer) ListenUnix(sock string) error {
	l, err := net.Listen("unix", sock)
	if err != nil {
		return err
	}
	go p.accept(l)
	return nil
}

func (p *PwxServer) HandleWs(conn *websocket.Conn) {
	t := newTransport(p.handlers)
	_ = t.readWrite(newWsReadWriter(conn))
}

func (p *PwxServer) accept(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		go p.handleConnection(conn)
	}
}

func (p *PwxServer) handleConnection(conn net.Conn) {
	t := newTransport(p.handlers)
	_ = t.readWrite(conn)
}
