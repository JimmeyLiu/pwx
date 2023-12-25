package transport

import (
	"errors"
	"github.com/gorilla/websocket"
)

type wsReadWriter struct {
	remain []byte
	conn   *websocket.Conn
}

func newWsReadWriter(conn *websocket.Conn) *wsReadWriter {
	return &wsReadWriter{
		remain: make([]byte, 0),
		conn:   conn,
	}
}

func (w *wsReadWriter) Read(buf []byte) (int, error) {
	bufSize := len(buf)
	idx := 0
	if len(w.remain) > 0 {
		for i, b := range w.remain {
			buf[i] = b
			idx++
			if idx >= bufSize {
				w.remain = w.remain[idx:]
				return idx, nil
			}
		}
	}
	messageType, p, err := w.conn.ReadMessage()
	if err != nil {
		return -1, err
	}
	if messageType != websocket.BinaryMessage {
		return -1, errors.New("bad message type")
	}

	for _, b := range p {
		if idx >= bufSize {
			w.remain = append(w.remain, b)
		} else {
			buf[idx] = b
			idx++
		}
	}

	return idx, nil
}

func (w *wsReadWriter) Write(p []byte) (int, error) {
	err := w.conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
