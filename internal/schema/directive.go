package schema

import (
	"errors"
	"fmt"
)

type Directive interface {
	Name() string
	Description() string
	Locations() []DirectiveLocation
	Args() []InputValue
}

var _ Directive = (*directive)(nil)

type DirectiveLocation string

const (
	DirectiveLocationQuery                DirectiveLocation = "QUERY"
	DirectiveLocationMutation                               = "MUTATION"
	DirectiveLocationSubscripiom                            = "SUBSCRIPTION"
	DirectiveLocationField                                  = "FIELD"
	DirectiveLocationFragmentDefinitionn                    = "FRAGMENT_DEFINITION"
	DirectiveLocationFragmentSpread                         = "FRAGMENT_SPREAD"
	DirectiveLocationInlineFragment                         = "INLINE_FRAGMENT"
	DirectiveLocationSchema                                 = "SCHEMA"
	DirectiveLocationScalar                                 = "SCALAR"
	DirectiveLocationObject                                 = "OBJECT"
	DirectiveLocationFieldDefinition                        = "FIELD_DEFINITION"
	DirectiveLocationArgumentDefinition                     = "ARGUMENT_DEFINITION"
	DirectiveLocationInterface                              = "INTERFACE"
	DirectiveLocationUnion                                  = "UNION"
	DirectiveLocationEnum                                   = "ENUM"
	DirectiveLocationEnumValue                              = "ENUM_VALUE"
	DirectiveLocationInputObject                            = "INPUT_OBJECT"
	DirectiveLocationInputFieldDefinition                   = "INPUT_FIELD_DEFINITION"
)

type DirectiveConfig struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Locations   []DirectiveLocation `json:"locations"`
	Args        []*InputValueConfig `json:"args,omitempty"`
}

func newDirective(schema *schema, cfg *DirectiveConfig) (*directive, error) {
	if schema == nil {
		return nil, errors.New("missing schema")
	}
	if cfg == nil {
		return nil, errors.New("missing config")
	}
	if cfg.Name == "" {
		return nil, errors.New("missing name")
	}
	directive := &directive{
		name:        cfg.Name,
		description: cfg.Description,
		locations:   cfg.Locations,
	}
	idx := map[string]bool{}
	for _, argCfg := range cfg.Args {
		arg, err := newInputValue(schema, argCfg)
		if err != nil {
			return nil, fmt.Errorf("in directive \"%s\": %s", directive.name, err)
		}
		if idx[arg.name] {
			return nil, fmt.Errorf(
				"in directive \"%s\": argument \"%s\" declared more than once",
				directive.name,
				arg.name,
			)
		}
		directive.args = append(directive.args, arg)
	}
	return directive, nil
}

type directive struct {
	name        string
	description string
	locations   []DirectiveLocation
	args        []InputValue
}

func (d *directive) Name() string                   { return d.name }
func (d *directive) Description() string            { return d.description }
func (d *directive) Locations() []DirectiveLocation { return d.locations }
func (d *directive) Args() []InputValue             { return d.args }
