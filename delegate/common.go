package delegate

import (
	"github.com/hashicorp/go-plugin"
	"github.com/tliron/kutil/logging"
)

var log = logging.GetLogger("khutulun.plugin")

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "khutulun",
}
