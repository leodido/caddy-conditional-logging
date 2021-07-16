package conditionallog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/logging"
	jsonselect "github.com/leodido/caddy-jsonselect-encoder"
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
	zapcore.Encoder       `json:"-"`
	zapcore.EncoderConfig `json:"-"`

	EncRaw    json.RawMessage `json:"encoder,omitempty" caddy:"namespace=caddy.logging.encoders inline_key=format"`
	Exprs     map[string][]expression
	Logger    *zap.Logger
	Formatter string
}

func (ce ConditionalEncoder) Clone() zapcore.Encoder {
	ret := ConditionalEncoder{
		Encoder:       ce.Encoder.Clone(),
		EncoderConfig: ce.EncoderConfig,
		Exprs:         ce.Exprs,
		Logger:        ce.Logger,
		Formatter:     ce.Formatter,
	}
	return ret
}

func (ce ConditionalEncoder) EncodeEntry(e zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	// Clone the original encoder to be sure we don't mess up it
	enc := ce.Encoder.Clone()
	if ce.Formatter == "console" {
		// todo > grab keys (eg. "msg") from the EncoderConfig
		// todo > set values according to line_ending, time_format, level_format
		// todo > duration_format too?
		enc.AddString(ce.LevelKey, e.Level.String())
		enc.AddTime(ce.TimeKey, e.Time)
		enc.AddString(ce.NameKey, e.LoggerName)
		enc.AddString(ce.MessageKey, e.Message)
		// todo > caller, stack
	} else if ce.Formatter == "jsonselect" {
		jsonEncoder, ok := ce.Encoder.(jsonselect.JSONSelectEncoder)
		if !ok {
			return nil, fmt.Errorf("unexpected encoder type %T", ce.Encoder)
		}
		enc = jsonEncoder.Encoder
	}

	// Store the logging encoder's buffer
	buf, err := enc.EncodeEntry(e, fields)
	if err != nil {
		return buf, err
	}
	data := buf.Bytes()

	// Strip non JSON-like prefix from the data buffer when it comes from a non JSON encoder
	if pos := bytes.Index(data, []byte(`{"`)); ce.Formatter == "console" && pos != -1 {
		data = data[pos:]
	}

	results := make(map[string][]bool)
	for key, expressions := range ce.Exprs {
		// Look for the field in the log entry
		path := strings.Split(key, ">")
		val, typ, _, err := jsonparser.Get(data, path...)
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
				// unreachable code because of the check on error above
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
		// Using the original (wrapped) encoder for output
		return ce.Encoder.EncodeEntry(e, fields)
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

func (ce *ConditionalEncoder) Provision(ctx caddy.Context) error {
	// Store the logger
	ce.Logger = ctx.Logger(ce)

	if ce.EncRaw == nil {
		ce.Encoder, ce.Formatter = newDefaultProductionLogEncoder(true)

		ctx.Logger(ce).Warn("fallback to a default production logging encoder")
		return nil
	}

	val, err := ctx.LoadModule(ce, "EncRaw")
	if err != nil {
		return fmt.Errorf("loading fallback encoder module: %v", err)
	}
	switch v := val.(type) {
	case *logging.JSONEncoder:
		ce.EncoderConfig = v.LogEncoderConfig.ZapcoreEncoderConfig()
	case *logging.ConsoleEncoder:
		ce.EncoderConfig = v.LogEncoderConfig.ZapcoreEncoderConfig()
	case *jsonselect.JSONSelectEncoder:
		ce.EncoderConfig = v.LogEncoderConfig.ZapcoreEncoderConfig()
	default:
		return fmt.Errorf("unsupported encoder type %T", v)
	}
	ce.Encoder = val.(zapcore.Encoder)

	return nil
}

func newDefaultProductionLogEncoder(colorize bool) (zapcore.Encoder, string) {
	encCfg := zap.NewProductionEncoderConfig()
	if term.IsTerminal(int(os.Stdout.Fd())) {
		// if interactive terminal, make output more human-readable by default
		encCfg.EncodeTime = func(ts time.Time, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(ts.UTC().Format("2006/01/02 15:04:05.000"))
		}
		if colorize {
			encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		return zapcore.NewConsoleEncoder(encCfg), "console"
	}
	return zapcore.NewJSONEncoder(encCfg), "json"
}

// Interface guards
var (
	_ zapcore.Encoder       = (*ConditionalEncoder)(nil)
	_ caddy.Provisioner     = (*ConditionalEncoder)(nil)
	_ caddyfile.Unmarshaler = (*ConditionalEncoder)(nil)
)
