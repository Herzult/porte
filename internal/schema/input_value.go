package schema

import (
	"errors"
	"fmt"
)

type InputValue interface {
	Name() string
	Description() string
	Type() Type
	DefaultValue() string
}

var _ InputValue = (*inputValue)(nil)

type InputValueConfig struct {
	Name         string
	Description  string
	Type         *TypeRefConfig
	DefaultValue string
}

func newInputValue(schema *schema, cfg *InputValueConfig) (*inputValue, error) {
	if schema == nil {
		return nil, errors.New("missing schema")
	}
	if cfg == nil {
		return nil, errors.New("missing config")
	}
	if cfg.Name == "" {
		return nil, errors.New("missing name")
	}
	t, err := newTypeRef(schema, cfg.Type)
	if err != nil {
		return nil, fmt.Errorf("invalid input value type: %s", err)
	}
	return &inputValue{
		name:         cfg.Name,
		description:  cfg.Description,
		typ:          t,
		defaultValue: cfg.DefaultValue,
	}, nil
}

type inputValue struct {
	name         string
	description  string
	typ          *typRef
	defaultValue string
}

func (i *inputValue) Name() string         { return i.name }
func (i *inputValue) Description() string  { return i.description }
func (i *inputValue) Type() Type           { return i.typ }
func (i *inputValue) DefaultValue() string { return i.defaultValue }
