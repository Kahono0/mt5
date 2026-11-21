package ec

import "fmt"

// interface for all models that can be parsed from a byte slice
type Parsable interface {
	Fields() []Field

	FromValues([]interface{}) any
}

func ParseToValue(m Parsable, body []byte) (interface{}, error) {
	values, err := Parse(body, m.Fields(), 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %T: %w", m, err)
	}

	return m.FromValues(values), nil
}

// generic function to parse a byte slice into a struct that implements the Parsable interface
func ParseGen[T Parsable](body []byte, m T) (T, error) {
	values, err := Parse(body, m.Fields(), 0)
	if err != nil {
		return m, fmt.Errorf("failed to parse %T: %w", m, err)
	}

	return m.FromValues(values).(T), nil
}
