package transport

import (
	"errors"
	"fmt"
	"github.com/JimmeyLiu/pwx/pkg/codec"
	"io"
	"log/slog"
	"strconv"
	"time"
)

type pwxTransport struct {
	rw        io.ReadWriter
	pending   map[string]chan Response
	streaming map[string]func(resp Response)
	handlers  map[string]RequestHandler
}

var codec0 = codec.NewCodec()

func newTransport(handlers map[string]RequestHandler) pwxTransport {
	cli := pwxTransport{
		pending:   make(map[string]chan Response),
		streaming: make(map[string]func(resp Response)),
		handlers:  handlers,
	}
	return cli
}

func (c *pwxTransport) doRequest(request Request, seconds int) Response {
	m, err := c.send(request, false)
	mid := codec.MidToString(m)

	if err != nil {
		slog.Error("send request error: " + err.Error())
		return Response{
			Status: 600,
			Mid:    mid,
		}
	}

	c.pending[mid] = make(chan Response)
	go timeout(seconds, func() {
		ch := c.pending[mid]
		if ch != nil {
			delete(c.pending, mid)
			ch <- Response{
				Status: 408, //超时
				Mid:    mid,
			}
		}
	})
	select {
	case resp := <-c.pending[mid]:
		return resp
	}
}

func (c *pwxTransport) doRequestStream(request Request, seconds int, callback func(resp Response)) {
	m, err := c.send(request, true)
	mid := codec.MidToString(m)

	if err != nil {
		callback(Response{
			Status: 600,
			Mid:    mid,
		})
		return
	}
	c.streaming[mid] = callback
	go timeout(seconds, func() {
		ch := c.streaming[mid]
		if ch != nil {
			delete(c.streaming, mid)
			callback(Response{
				Status: 408, //超时
				Mid:    mid,
			})
		}
	})
}

func timeout(seconds int, callback func()) {
	if seconds <= 0 {
		seconds = 3
	}
	d, _ := time.ParseDuration(fmt.Sprintf("%ds", seconds))
	time.AfterFunc(d, callback)
}

func (c *pwxTransport) signal(request Request) error {
	_, err := c.send(request, false)
	return err
}

func (c *pwxTransport) send(request Request, steam bool) ([]byte, error) {
	if c.rw == nil {
		return nil, errors.New("not connected")
	}

	msgType := codec.Request
	if steam {
		msgType = codec.RequestStream
	}

	frame := codec.Frame{
		Mid:       codec.NewMid(),
		MsgType:   msgType,
		StartLine: request.Path,
		Headers:   request.Headers,
		Body:      request.Body,
	}
	err := c.write(frame)
	if err != nil {
		return nil, err
	}
	return frame.Mid, nil
}

func (c *pwxTransport) reply(response Response) {
	frame := codec.Frame{
		MsgType:   codec.Response,
		Mid:       codec.StrToMid(response.Mid),
		StartLine: strconv.Itoa(response.Status),
		Headers:   response.Headers,
		Body:      response.Body,
	}
	_ = c.write(frame)
}

func (c *pwxTransport) write(frame codec.Frame) error {
	b, err := codec0.Encode(frame)
	if err != nil {
		return err
	}
	_, err = c.rw.Write(b)
	if err != nil {
		return err
	}
	return nil
}

func (c *pwxTransport) readWrite(rw io.ReadWriter) error {
	defer c.close()
	c.rw = rw
	return codec0.ReadFrame(rw, func(frame *codec.Frame) {
		if frame.MsgType == codec.Signal {
			h := c.handlers[frame.StartLine]
			if h != nil {
				_ = h.Handle(Request{
					Path:    frame.StartLine,
					Headers: frame.Headers,
					Body:    frame.Body,
				})
			}
		} else if frame.MsgType == codec.Request || frame.MsgType == codec.RequestStream {
			h := c.handlers[frame.StartLine]
			mid := codec.MidToString(frame.Mid)
			if h != nil {
				go func() {
					req := Request{
						Mid:     mid,
						Path:    frame.StartLine,
						Headers: frame.Headers,
						Body:    frame.Body,
					}
					if frame.MsgType == codec.Request {
						resp := h.Handle(req)
						resp.Mid = mid
						c.reply(resp)
					} else {
						h.HandleStream(req, func(resp Response) {
							resp.Mid = mid
							c.reply(resp)
						})
					}
				}()
			} else {
				c.reply(Response{
					Status: 404,
					Mid:    mid,
				})
			}
		} else if frame.MsgType == codec.Response {
			status, err := strconv.Atoi(frame.StartLine)
			if err != nil {
				status = 500
			}

			mid := codec.MidToString(frame.Mid)
			resp := Response{
				Status:  status,
				Headers: frame.Headers,
				Body:    frame.Body,
				Mid:     mid,
			}

			ch := c.pending[mid]

			if ch != nil {
				ch <- resp
			} else {
				stream := c.streaming[mid]
				if stream != nil {
					if status >= 200 {
						delete(c.streaming, mid)
					}
					stream(resp)
				}
			}
		}
	})
}

func (c *pwxTransport) close() {
	for k, ch := range c.pending {
		ch <- Response{
			Status: 600,
			Mid:    k,
		}
		delete(c.pending, k)
	}
	for k, v := range c.streaming {
		v(Response{
			Status: 600,
			Mid:    k,
		})
		delete(c.streaming, k)
	}
}
