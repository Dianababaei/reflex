package outbound

import (
	"github.com/xtls/xray-core/common/protocol"
)

// Config represents outbound configuration (matches proto definition)
type Config struct {
	Vnext []*protocol.ServerEndpoint
}
