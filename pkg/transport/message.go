package transport

import (
	"github.com/JimmeyLiu/pwx/pkg/codec"
	"strconv"
)

type Request struct {
	Path    string
	Mid     string
	Headers map[string]string
	Body    []byte
}

type Response struct {
	Status  int //状态码，含义和http status code一致
	Mid     string
	Headers map[string]string
	Body    []byte
}

func (resp Response) isSuccess() bool {
	return resp.Status >= 200 && resp.Status < 300
}

func (resp Response) isRunning() bool {
	return resp.Status < 200
}

func (resp Response) isFailed() bool {
	return resp.Status >= 300
}

type RequestHandler interface {
	Path() string
	Handle(req Request) Response
	HandleStream(req Request, reply func(resp Response))
}

func EncodeRequest(request Request) ([]byte, error) {
	frame := codec.Frame{
		Mid:       codec.NewMid(),
		MsgType:   codec.Request,
		StartLine: request.Path,
		Headers:   request.Headers,
		Body:      request.Body,
	}
	return codec0.Encode(frame)
}

func EncodeResponse(response Response) ([]byte, error) {
	frame := codec.Frame{
		MsgType:   codec.Response,
		Mid:       codec.StrToMid(response.Mid),
		StartLine: strconv.Itoa(response.Status),
		Headers:   response.Headers,
		Body:      response.Body,
	}
	return codec0.Encode(frame)
}
