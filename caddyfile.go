// Copyright 2021 Leonardo Di Donato
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package conditionallog

import (
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"go.uber.org/zap/zapcore"
)

// UnmarshalCaddyfile sets up the module form Caddyfile tokens.
//
// Syntax:
// if {
//     "<expression>"
// } [<encoder>]
//
// The <expression> must be on a single line.
// Refer to `lang.Lang` for its syntax.
//
// The <encoder> can be one of `json`, `jsonselector`, `console`.
// In case no <encoder> is specified, one between `json` and `console` is set up depending
// on the current environment.
func (ce *ConditionalEncoder) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if d.Next() {
		if d.Val() != moduleName {
			return d.Errf("expecting %s (%T) subdirective", moduleID, ce)
		}
		var expression string
		if !d.Args(&expression) {
			return d.Errf("%s (%T) requires an expression", moduleID, ce)
		}

		ce.Expr = expression
	}

	if !d.Next() {
		return nil
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
