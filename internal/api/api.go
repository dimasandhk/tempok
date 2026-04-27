package api

import "time"

type HandshakeRequest struct {
	Secret string `json:"secret"`
	Type   string `json:"type"` // "tunnel" or "api"
}

type HandshakeResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// RPC Args and Replies

type ListArgs struct{}
type ListReply struct {
	Tunnels []TunnelInfo
}

type TunnelInfo struct {
	ID         string
	ExpiresAt  time.Time
	IsForever  bool
	PublicPort int
}

type ExtendArgs struct {
	ID       string
	Duration time.Duration
}
type ExtendReply struct {
	Error string
}

type PersistArgs struct {
	ID string
}
type PersistReply struct {
	Error string
}
