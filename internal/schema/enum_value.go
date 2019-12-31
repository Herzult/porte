package schema

import "errors"

type EnumValue interface {
	Name() string
	Description() string
	IsDeprecated() bool
	DeprecationReason() string
}

var _ EnumValue = (*enumValue)(nil)

type EnumValueConfig struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	IsDeprecated      bool   `json:"isDeprecated,omitempty"`
	DeprecationReason string `json:"deprecationReason,omitempty"`
}

func newEnumValue(cfg *EnumValueConfig) (*enumValue, error) {
	if cfg == nil {
		return nil, errors.New("missing config")
	}
	if cfg.Name == "" {
		return nil, errors.New("missing name for enum value")
	}
	return &enumValue{
		name:              cfg.Name,
		description:       cfg.Description,
		isDeprecated:      cfg.IsDeprecated,
		deprecationReason: cfg.DeprecationReason,
	}, nil
}

type enumValue struct {
	name              string
	description       string
	isDeprecated      bool
	deprecationReason string
}

func (e *enumValue) Name() string              { return e.name }
func (e *enumValue) Description() string       { return e.description }
func (e *enumValue) IsDeprecated() bool        { return e.isDeprecated }
func (e *enumValue) DeprecationReason() string { return e.deprecationReason }
