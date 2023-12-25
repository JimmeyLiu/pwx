package codec

import "errors"

type PwxV1 struct {
}

const V1 = byte(1)

func (p *PwxV1) Version() byte {
	return V1
}

func (p *PwxV1) Encode(frame Frame) ([]byte, error) {

	startLine := []byte(frame.StartLine)
	startLineLen := len(startLine)
	midLen := len(frame.Mid)
	if midLen > 255 {
		return nil, errors.New("mid too large")
	}
	headBytes := p.encodeHeads(frame.Headers)
	headLen := len(headBytes)
	if headLen > 1000 {
		return nil, errors.New("headers too large")
	}

	totalLen := 5 + startLineLen + midLen + headLen + len(frame.Body) // 5 = startLineLen -> 1 + msgType -> 1 + midLen ->1 + headLen ->2

	b := make([]byte, 0)
	b = append(b, MAGIC...)                  //magic 3
	b = append(b, V1)                        //version 1
	b = append(b, int32ToBytes(totalLen)...) //total len 4
	b = append(b, frame.MsgType)             //MsgType len 1
	b = append(b, byte(startLineLen))        // StartLine len 1
	b = append(b, startLine...)              // slen
	b = append(b, byte(midLen))              // Mid len 1
	b = append(b, frame.Mid...)              // midLen
	b = append(b, int16ToBytes(headLen)...)  //Headers len 2
	b = append(b, headBytes...)              //headLen
	b = append(b, frame.Body...)
	return b, nil
}

func (p *PwxV1) Decode(data []byte) (*Frame, error) {
	frame := &Frame{}
	total := bytesToInt32(data[4:8])
	if len(data) != total+8 {
		return nil, errors.New("bad frame")
	}

	frame.MsgType = data[8]

	startLineLen := int(data[9])
	var idx = 10
	var next = 10 + startLineLen
	if startLineLen > 0 {
		frame.StartLine = string(data[idx:next])
	}

	midLen := int(data[next])
	if midLen > 0 {
		idx = next + 1
		next = idx + midLen
		frame.Mid = data[idx:next]
	}

	idx = next
	next = idx + 2
	headLen := bytesToInt16(data[idx:next])
	if headLen > 0 {
		idx = next
		next = idx + headLen
		frame.Headers = p.decodeHeads(data[idx:next])
	}
	frame.Body = data[next:]
	return frame, nil
}

func (p *PwxV1) encodeHeads(head map[string]string) []byte {
	headBytes := make([]byte, 0)
	if head == nil || len(head) == 0 {
		return headBytes
	}
	for k, v := range head {
		if k == "" || v == "" {
			continue
		}
		h := p.encodeHead(k, v)
		if h != nil {
			headBytes = append(headBytes, h...)
		}
	}
	return headBytes
}

func (p *PwxV1) encodeHead(k, v string) []byte {
	head := make([]byte, 0)
	kb := []byte(k)
	vb := []byte(v)

	head = append(head, byte(len(kb)))
	head = append(head, byte(len(vb)))
	head = append(head, kb...)
	head = append(head, vb...)
	return head
}

func (p *PwxV1) decodeHeads(d []byte) map[string]string {
	heads := make(map[string]string)
	for len(d) > 0 {
		klen := d[0]
		vlen := d[1]
		next := 2 + klen + vlen
		k := string(d[2:klen])
		v := string(d[2+klen : next])
		heads[k] = v
		d = d[next:]
	}
	return heads
}
