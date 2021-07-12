package conditionallog

import (
	"github.com/caddyserver/caddy/v2"
)

func init() {
	caddy.RegisterModule(ConditionalLogger{})
}

type ConditionalLogger struct {
}

func (ConditionalLogger) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "leodido.conditional_log",
		New: func() caddy.Module {
			return new(ConditionalLogger)
		},
	}
}
