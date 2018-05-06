package model

const (
	UNIDENTIFIED = "UNIDENTIFIED"
	HELLO        = "HELLO"
	MEMBERSHIP   = "MEMBERSHIP"
)

var FeiwuMessageTypes = map[byte]string{
	0x00: UNIDENTIFIED,
	0x01: HELLO,
	0x02: MEMBERSHIP,
}
