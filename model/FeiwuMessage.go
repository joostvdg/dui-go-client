package model

import "time"

// FeiwuMessage information about a message, including the payload.
type FeiwuMessage struct {
	Message       string             `json:"Message"`
	MessageDigest []byte             `json:"MessageDigest"`
	MessageOrigin FeiwuMessageOrigin `json:"HostName"`
	MessageType   string             `json:"MessageType"`
	ReceivedRaw   time.Time          `json:"ReceivedRaw"`
	Received      string             `json:"Received"`
}
