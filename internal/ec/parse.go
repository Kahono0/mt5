package ec

import (
	"fmt"
	"reflect"
	"sync"
)

var parsableSchemas sync.Map // map[reflect.Type]*Schema

// Parsable implementations are expected to return the same field layout for
// every value of a given concrete type.
func schemaFor(m Parsable) *Schema {
	t := reflect.TypeOf(m)
	if schema, ok := parsableSchemas.Load(t); ok {
		return schema.(*Schema)
	}

	schema, _ := parsableSchemas.LoadOrStore(t, CompileSchema(m.Fields()))
	return schema.(*Schema)
}

// interface for all models that can be parsed from a byte slice
type Parsable interface {
	Fields() []Field

	FromValues([]interface{}) any
}

// ByteParser is an optional fast path for models with a handwritten decoder.
// Implementations should populate the receiver directly from the wire bytes.
type ByteParser interface {
	ParseFromBytes([]byte) error
}

func ParseToValue(m Parsable, body []byte) (interface{}, error) {
	if parser, ok := m.(ByteParser); ok {
		if err := parser.ParseFromBytes(body); err != nil {
			return nil, fmt.Errorf("failed to parse %T: %w", m, err)
		}
		return m, nil
	}

	values, err := schemaFor(m).Parse(body, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %T: %w", m, err)
	}

	return m.FromValues(values), nil
}

// generic function to parse a byte slice into a struct that implements the Parsable interface
func ParseGen[T Parsable](body []byte, m T) (T, error) {
	if parser, ok := any(m).(ByteParser); ok {
		if err := parser.ParseFromBytes(body); err != nil {
			return m, fmt.Errorf("failed to parse %T: %w", m, err)
		}
		return m, nil
	}

	values, err := schemaFor(m).Parse(body, 0)
	if err != nil {
		return m, fmt.Errorf("failed to parse %T: %w", m, err)
	}

	return m.FromValues(values).(T), nil
}
