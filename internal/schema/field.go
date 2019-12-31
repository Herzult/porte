package schema

import (
	"errors"
	"fmt"
)

type Field interface {
	Name() string
	Description() string
	Args() []InputValue
	Arg(string) InputValue
	Type() Type
	IsDeprecated() bool
	DeprecationReason() string
}

var _ Field = (*field)(nil)

type FieldConfig struct {
	Name              string
	Description       string
	Args              []*InputValueConfig
	Type              *TypeRefConfig
	IsDeprecated      bool
	DeprecationReason string
}

func newField(schema *schema, cfg *FieldConfig) (*field, error) {
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
		return nil, fmt.Errorf("invalid type in field \"%s\": %s", cfg.Name, err)
	}
	f := &field{
		name:              cfg.Name,
		description:       cfg.Description,
		isDeprecated:      cfg.IsDeprecated,
		deprecationReason: cfg.DeprecationReason,
		typ:               t,
		argsMap:           map[string]*inputValue{},
	}
	for _, argCfg := range cfg.Args {
		arg, err := newInputValue(schema, argCfg)
		if err != nil {
			return nil, fmt.Errorf("invalid arg config: %s", err)
		}
		if _, ok := f.argsMap[arg.name]; ok {
			return nil, fmt.Errorf("arg \"%s\" defined more than once", arg.name)
		}
		f.argsMap[arg.name] = arg
		f.args = append(f.args, arg)
	}
	return f, nil
}

type field struct {
	name              string
	description       string
	args              []InputValue
	argsMap           map[string]*inputValue
	typ               Type
	isDeprecated      bool
	deprecationReason string
}

func (f *field) Name() string               { return f.name }
func (f *field) Description() string        { return f.description }
func (f *field) Args() []InputValue         { return f.args }
func (f *field) Arg(name string) InputValue { return f.argsMap[name] }
func (f *field) Type() Type                 { return f.typ }
func (f *field) IsDeprecated() bool         { return f.isDeprecated }
func (f *field) DeprecationReason() string  { return f.deprecationReason }
