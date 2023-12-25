package codec

func NewMid() []byte {
	return []byte("m1")
}

func MidToString(mid []byte) string {
	return string(mid)
}

func StrToMid(mid string) []byte {
	return []byte(mid)
}
