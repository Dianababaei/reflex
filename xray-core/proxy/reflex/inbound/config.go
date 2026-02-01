package inbound

import (
	"github.com/xtls/xray-core/common/protocol"
)

// Fallback represents fallback configuration (matches proto definition)
type Fallback struct {
	Name string
	Alpn string
	Path string
	Type string
	Dest string
	Xver uint64
}

// Config represents inbound configuration (matches proto definition)
type Config struct {
	Clients   []*protocol.User
	Fallbacks []*Fallback
}
