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
	"bytes"
	"fmt"
)

type Op func([]byte, []byte) bool

var operands = map[string]Op{
	"ne": func(val1, val2 []byte) bool { return bytes.Compare(val1, val2) != 0 },
	"eq": func(val1, val2 []byte) bool { return bytes.Compare(val1, val2) == 0 },
	"sw": func(val1, val2 []byte) bool { return bytes.HasPrefix(val1, val2) },
}

type expression struct {
	Op string
	Lh string
	Rh []byte
}

func makeExpression(op, lh string, rh string) (expression, error) {
	// Is operator valid?
	_, ok := operands[op]
	if !ok {
		return expression{}, fmt.Errorf("unknown operator %q", op)
	}

	return expression{op, lh, []byte(rh)}, nil
}

func (e expression) Evaluate(value []byte) bool {
	fx, ok := operands[e.Op]
	if !ok {
		// Unreachable when `e` has been built with `makeExpression`
		return false
	}

	return fx(value, e.Rh)
}
