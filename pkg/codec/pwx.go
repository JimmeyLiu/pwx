package codec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	Signal        byte = 0 //信号消息，无需等待响应
	Request       byte = 1 //请求消息
	RequestStream byte = 2
	Response      byte = 3 //请求响应
)

type Frame struct {
	Mid       []byte            //消息ID
	MsgType   byte              //消息类型, 0 signal , 1 request ,2 response , 3 stream reply
	StartLine string            //消息路径，类似http path
	Headers   map[string]string //头信息
	Body      []byte            //消息体
}

type Pwx interface {
	Version() byte
	Encode(frame Frame) ([]byte, error)
	Decode(data []byte) (*Frame, error)
}

var MAGIC = []byte("PWX")

type PwxCodec struct {
	v1      Pwx
	current Pwx
}

func NewCodec() *PwxCodec {
	v1 := &PwxV1{}
	return &PwxCodec{
		v1:      v1,
		current: v1,
	}
}

func (p *PwxCodec) Encode(frame Frame) ([]byte, error) {
	return p.current.Encode(frame)
}

func (p *PwxCodec) Decode(data []byte) (*Frame, error) {
	if len(data) <= 8 {
		return nil, errors.New("bad frame")
	}
	if !bytes.Equal(MAGIC, data[:3]) {
		return nil, errors.New("bad MAGIC")
	}
	version := data[3]
	if version == 1 {
		return p.v1.Decode(data)
	}
	return nil, errors.New(fmt.Sprintf("un support version %d", version))
}

func (p *PwxCodec) ReadFrame(r io.Reader, callback func(frame *Frame)) error {
	buf := make([]byte, 1024*512)
	data := make([]byte, 0)
	for {
		n, err := r.Read(buf)
		if err != nil {
			return err
		}
		data = append(data, buf[:n]...)
		l := len(data)
		if l < 8 {
			continue
		}
		total := bytesToInt32(data[4:8])
		frameLen := total + 8
		if l < frameLen {
			continue
		}
		frame, err := p.Decode(data[:frameLen])
		if err != nil {
			return err
		}
		//回调到业务层，业务层处理出错这里不用管
		callback(frame)
		data = data[frameLen:]
	}
}

func int16ToBytes(i int) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(i))
	return b
}

func int32ToBytes(i int) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(i))
	return b
}

func bytesToInt32(b []byte) int {
	return int(binary.BigEndian.Uint32(b))
}

func bytesToInt16(b []byte) int {
	return int(binary.BigEndian.Uint16(b))
}
