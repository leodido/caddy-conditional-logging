package conditionallog

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func init() {
	caddy.RegisterModule(ConditionalOut{})
}

type ConditionalOut struct {
	Writer                string          `json:"writer,omitempty"`
	ConditionalEncoderRaw json.RawMessage `json:"-" caddy:"namespace=caddy.logging.encoders inline_key=format"`
}

// WriterKey returns a unique key representing the receiving conditional logger.
func (co ConditionalOut) WriterKey() string {
	return "conditional:" + co.Writer
}

// OpenWriter returns os.Stdout or os.Stderr.
func (co ConditionalOut) OpenWriter() (io.WriteCloser, error) {
	fmt.Println("==> OPENWRITER", co.Writer)
	// Proxy to builtin writers
	switch co.Writer {
	case "stdout":
		return (caddy.StdoutWriter{}).OpenWriter()
	case "stderr":
		fallthrough
	default:
		return (caddy.StderrWriter{}).OpenWriter()
	}
}

// String returns a string representing the receivine conditional logger.
func (ConditionalOut) String() string {
	return "conditional"
}

func (ConditionalOut) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "caddy.logging.writers.conditional", // see https://github.com/caddyserver/caddy/blob/ef7f15f3a42474319e2db0dff6720d91c153f0bf/caddyconfig/httpcaddyfile/builtins.go#L701
		New: func() caddy.Module {
			return new(ConditionalOut)
		},
	}
}

func (co *ConditionalOut) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	fmt.Println("==> UNMARSHAL")
	for d.Next() {
		if !d.NextArg() {
			return d.ArgErr()
		}
		co.Writer = d.Val()
		// todo > check is stdout or stderr etc.
		if d.NextArg() {
			return d.ArgErr()
		}

		// todo > continue
	}
	return nil
}

func (co *ConditionalOut) Provision(ctx caddy.Context) error {
	fmt.Println("==> PROVISION")
	// Use the builtin high-performance logger
	// ce.logger = ctx.Logger(ce)

	if co.ConditionalEncoderRaw == nil {
		val, err := ctx.LoadModule(co, "ConditionalEncoderRaw")
		fmt.Println(val)
		fmt.Println(err)
	}

	if co.Writer == "" {
		co.Writer = "stderr"
	}

	return nil
}

// Interface guards
var (
	_ caddy.WriterOpener = (*ConditionalOut)(nil)
)
