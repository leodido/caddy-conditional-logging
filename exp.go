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
