package conditionallog

import (
	"fmt"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"go.uber.org/zap/zapcore"
)

// UnmarshalCaddyfile sets up the module form Caddyfile tokens.
//
// Syntax:
// if {
//     <expression>
// } [<encoder>]
//
// The <expression> syntax is <field> <operator> <value>.
// More <expression>s can be defined.
// They will be evaluated in OR.
//
// The <encoder> can be one of `json`, `jsonselector`, `console`.
// In case no <encoder> is specified, one between `json` and `console` is set up depending
// on the current environment.
func (ce *ConditionalEncoder) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	rest := 0
	ntokens := 0
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
			ntokens++
			args := d.RemainingArgs()
			if len(args) == 2 {
				ntokens += 2
				operand := args[0]
				value := args[1]

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
				continue
			}

			// No more args: remove field from the token count an go outside the loop
			if len(args) == 0 {
				ntokens--
				break
			}

			// The only supported encoder that can have an argument is `jsonselect`
			// Error if leftover arguments are present
			if len(args) > 0 && field != "jsonselect" {
				return d.Errf("expecting <field> <operator> <value> tokens for %s (%T)", moduleID, ce)
			}

			// Keep track of the number of arguments to backtrack
			rest = len(args)
			// Do not include them in the expression tokens count
			ntokens -= rest

			// Backtrack to the `jsonselect` token
			// Adjust the number of arguments to backtrack accordingly
			// Exit the loop
			if len(args) > 0 && field == "jsonselect" {
				d.Prev()
				rest--
				break
			}
		}
	}

	// Double-check the expressions are well-formed
	if ntokens%3 != 0 {
		return d.Errf("expecting <field> <operator> <value> tokens for %s (%T)", moduleID, ce)
	}

	// Advance when a supported encoder is found
	switch d.Val() {
	case "json":
		fallthrough
	case "jsonselect":
		fmt.Println("FALL")
		fallthrough
	case "console":
		d.NextBlock(0)
	default:
		return nil
	}

	// Backtracking to before the <encoder> starts
	for i := 0; i <= rest; i++ {
		d.Prev()
	}

	// Delegate the parsing of the encoder to the encoder itself
	nextDispenser := d.NewFromNextSegment()
	if nextDispenser.Next() {
		moduleName := nextDispenser.Val()
		moduleID := "caddy.logging.encoders." + moduleName
		mod, err := caddyfile.UnmarshalModule(nextDispenser, moduleID)
		if err != nil {
			return err
		}
		enc, ok := mod.(zapcore.Encoder)
		if !ok {
			return d.Errf("module %s (%T) is not a zapcore.Encoder", moduleID, mod)
		}
		ce.EncRaw = caddyconfig.JSONModuleObject(enc, "format", moduleName, nil)
		ce.Formatter = moduleName
	}

	return nil
}
