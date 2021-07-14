package conditionallog

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"golang.org/x/term"
)

const (
	moduleName = "if"
	moduleID   = "caddy.logging.encoders." + moduleName
)

func init() {
	caddy.RegisterModule(ConditionalEncoder{})
}

type ConditionalEncoder struct {
	zapcore.Encoder `json:"-"`

	EncRaw json.RawMessage `json:"encoder,omitempty" caddy:"namespace=caddy.logging.encoders inline_key=format"`
	Exprs  map[string][]expression
	Logger *zap.Logger
}

func (ce ConditionalEncoder) Clone() zapcore.Encoder {
	ret := ConditionalEncoder{
		Encoder: ce.Encoder.Clone(),
		Exprs:   ce.Exprs,
	}
	return ret
}

func (ce ConditionalEncoder) EncodeEntry(e zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	// TODO > always use a JSON encoder to obtain the buffer

	// Store the logging encoder's buffer
	buf, err := ce.Encoder.EncodeEntry(e, fields)
	if err != nil {
		return buf, err
	}

	results := make(map[string][]bool)
	for key, expressions := range ce.Exprs {
		// Look for the field in the log entry
		path := strings.Split(key, ">")
		val, typ, _, err := jsonparser.Get(buf.Bytes(), path...)
		if err != nil {
			// Field not found, ignore the current expression
			// todo > warn?
			continue
		}

		// Evaluate the expressions regarding the current field
		for _, e := range expressions {
			// Switch on the actual type of the value
			switch typ {
			case jsonparser.NotExist:
				// todo > unreachable code because of the upper check on error, but warn anyway?
			case jsonparser.Number, jsonparser.String:
				results[key] = append(results[key], e.Evaluate(val))
			default:
				// todo > warn that we don't support this type
			}
		}
	}

	acc := false
	for _, res := range results {
		for _, r := range res {
			acc = acc || r
		}
	}

	if acc || len(results) == 0 {
		return buf, err
	}
	return nil, nil
}

func (ConditionalEncoder) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: moduleID, // see https://github.com/caddyserver/caddy/blob/ef7f15f3a42474319e2db0dff6720d91c153f0bf/caddyconfig/httpcaddyfile/builtins.go#L720
		New: func() caddy.Module {
			return new(ConditionalEncoder)
		},
	}
}

// todo >
// func (ce *ConditionalEncoder) Validate() error {}

// UnmarshalCaddyfile sets up the module form Caddyfile tokens.
//
// Syntax:
// condition {
//     <field> <operator> <value>
// }
func (ce *ConditionalEncoder) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if d.Next() {
		if d.Val() != moduleName {
			return d.Errf("expecting %s (%T) subdirective", moduleID, ce)
		}
		// Expecting a block opening
		if d.NextArg() {
			return d.ArgErr()
		}
		nconditions := 0
		// Parse block
		for d.NextBlock(0) {
			field := d.Val()
			if !d.NextArg() {
				if nconditions == 0 {
					return d.ArgErr()
				} else {
					d.Prev()
					break
				}
			}

			operand := d.Val()
			if !d.NextArg() {
				return d.ArgErr()
			}

			value := d.Val()
			if nconditions == 0 {
				ce.Exprs = make(map[string][]expression)
			}
			exp, err := makeExpression(operand, field, value)
			if err != nil {
				return d.Err(err.Error())
			}
			if _, ok := ce.Exprs[field]; !ok {
				ce.Exprs[field] = make([]expression, 0)
			}
			ce.Exprs[field] = append(ce.Exprs[field], exp)

			nconditions++
		}
	}

	if d.Next() {
		moduleName := d.Val()
		moduleID := "caddy.logging.encoders." + moduleName
		mod, err := caddyfile.UnmarshalModule(d, moduleID)
		if err != nil {
			return err
		}
		enc, ok := mod.(zapcore.Encoder)
		if !ok {
			return d.Errf("module %s (%T) is not a zapcore.Encoder", moduleID, mod)
		}
		ce.EncRaw = caddyconfig.JSONModuleObject(enc, "format", moduleName, nil)
	}

	return nil
}

func (ce *ConditionalEncoder) Provision(ctx caddy.Context) error {
	// Store the logger
	ce.Logger = ctx.Logger(ce)

	if ce.EncRaw == nil {
		ce.Encoder = newDefaultProductionLogEncoder(true)

		ctx.Logger(ce).Warn("fallback to a default production logging encoder")
		return nil
	}

	val, err := ctx.LoadModule(ce, "EncRaw")
	if err != nil {
		return fmt.Errorf("loading fallback encoder module: %v", err)
	}
	ce.Encoder = val.(zapcore.Encoder)

	return nil
}

func newDefaultProductionLogEncoder(colorize bool) zapcore.Encoder {
	encCfg := zap.NewProductionEncoderConfig()
	if term.IsTerminal(int(os.Stdout.Fd())) {
		// if interactive terminal, make output more human-readable by default
		encCfg.EncodeTime = func(ts time.Time, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(ts.UTC().Format("2006/01/02 15:04:05.000"))
		}
		if colorize {
			encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		return zapcore.NewConsoleEncoder(encCfg)
	}
	return zapcore.NewJSONEncoder(encCfg)
}

// Interface guards
var (
	_ zapcore.Encoder       = (*ConditionalEncoder)(nil)
	_ caddy.Provisioner     = (*ConditionalEncoder)(nil)
	_ caddyfile.Unmarshaler = (*ConditionalEncoder)(nil)
)
