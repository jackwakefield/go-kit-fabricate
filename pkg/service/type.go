package service

import (
	"fmt"
	"math/rand"
	"strings"

	parser "github.com/jackwakefield/go-parser"
)

type TypeValue interface {
	Name() string
	LocalName() string
	GlobalName() string
	Type() string
	IsContext() bool
	IsErr() bool
	ZeroValue() string
}

type typeValue struct {
	item *parser.GoType
	name string
}

func newTypeValue(item *parser.GoType) TypeValue {
	return &typeValue{
		item: item,
	}
}

func (t *typeValue) Name() string {
	if t.name == "" {
		t.name = t.item.Name
	}

	if t.name == "" {
		t.name = fmt.Sprintf("fab%x", rand.Uint64())
	}

	return t.name
}

func (t *typeValue) LocalName() string {
	name := t.Name()

	if len(name) > 0 {
		lowercase := strings.ToLower(string(name[0]))

		if len(name) == 1 {
			name = lowercase
		} else if len(name) >= 1 {
			name = lowercase + string(name[1:])
		}
	}

	return name
}

func (t *typeValue) GlobalName() string {
	return strings.Title(t.Name())
}

func (t *typeValue) Type() string {
	return t.item.Type
}

func (t *typeValue) IsContext() bool {
	// TODO: more robust type checking
	if t.Type() == "context.Context" {
		return true
	}
	return false
}

func (t *typeValue) IsErr() bool {
	isError := func(typeName string) bool {
		return typeName == "error"
	}

	if isError(t.Type()) {
		return true
	}
	if t.item.Underlying != "" {
		if isError(t.item.Underlying) {
			return true
		}
	}
	return false
}

func (t *typeValue) ZeroValue() string {
	zeroValue := func(typeName string) string {
		switch typeName {
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
			return "0"
		case "bool":
			return "false"
		case "string":
			return "\"\""
		default:
			return "nil"
		}
	}

	zero := zeroValue(t.Type())
	if zero == "nil" {
		if t.item.Underlying != "" {
			zero = zeroValue(t.item.Underlying)
		}
	}
	return zero
}
