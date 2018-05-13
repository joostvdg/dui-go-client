package model

import "time"

// FeiwuMessageOrigin information about origin of a message, such as hostname, ip.
type FeiwuMessageOrigin struct {
	HostName    string    `json:"HostName"`
	HostIP      string    `json:"HostIP"`
	ServerName  string    `json:"ServerName"`
	Role        string    `json:"Role"`
	LastSeenRaw time.Time `json:"LastSeenRaw"`
	LastSeen    string    `json:"LastSeen"`
}
