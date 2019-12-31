package schema

import (
	"errors"
	"fmt"
)

type Schema interface {
	QueryType() Type
	MutationType() Type
	SubscriptionType() Type
	Types() []Type
	Type(string) Type
	Directives() []Directive
	Directive(string) Directive
}

var _ Schema = (*schema)(nil)

// SchemaConfig represents a schema
type SchemaConfig struct {
	QueryType        *TypeRefConfig     `json:"queryType"`
	MutationType     *TypeRefConfig     `json:"mutationType,omitempty"`
	SubscriptionType *TypeRefConfig     `json:"subscriptionType,omitempty"`
	Types            []*TypeConfig      `json:"types"`
	Directives       []*DirectiveConfig `json:"directives"`
}

// NewSchema returns a new Schema instance from the given SchemaConfig
func NewSchema(cfg *SchemaConfig) (Schema, error) {
	schema := &schema{
		typesMap:         map[string]Type{},
		directivesMap:    map[string]Directive{},
		typesByInterface: map[string][]Type{},
	}

	// index fields
	for _, typCfg := range cfg.Types {
		typ, err := newType(schema, typCfg)
		if err != nil {
			return nil, fmt.Errorf("invalid config: %s", err)
		}
		if _, ok := schema.typesMap[typ.Name()]; ok {
			return nil, fmt.Errorf("type \"%s\" defined more than once", typ.Name())
		}
		schema.types = append(schema.types, typ)
		schema.typesMap[typ.Name()] = typ
	}

	// index directives
	for _, dirCfg := range cfg.Directives {
		dir, err := newDirective(schema, dirCfg)
		if err != nil {
			return nil, fmt.Errorf("found invalid directive: %s", err)
		}
		if _, ok := schema.directivesMap[dir.Name()]; ok {
			return nil, fmt.Errorf("directive \"%s\" defined more than once", dir.Name())
		}
		schema.directivesMap[dir.Name()] = dir
		schema.directives = append(schema.directives, dir)
	}

	// assign root types
	if cfg.QueryType != nil {
		typ, err := newTypeRef(schema, cfg.QueryType)
		if err != nil {
			return nil, fmt.Errorf("invalid config: query type: %s", err)
		}
		if _, ok := schema.typesMap[typ.name]; !ok && typ.name != "" {
			return nil, fmt.Errorf("invalid config: query type references non-existing type \"%s\"", typ.name)
		}
		if typ.Kind() != TypeKindObject {
			return nil, fmt.Errorf(
				"invalid config: query type must reference an OBJECT type, found %s type",
				typ.Kind(),
			)
		}
		schema.queryType = typ
	} else {
		return nil, errors.New("invalid config: no query type specified")
	}
	if cfg.MutationType != nil {
		typ, err := newTypeRef(schema, cfg.MutationType)
		if err != nil {
			return nil, fmt.Errorf("invalid config: mutation type: %s", err)
		}
		if _, ok := schema.typesMap[typ.name]; !ok && typ.name != "" {
			return nil, fmt.Errorf("invalid config: mutation type references non-existing type \"%s\"", typ.name)
		}
		if typ.Kind() != TypeKindObject {
			return nil, fmt.Errorf(
				"invalid config: mutation type must reference an OBJECT type, found %s type",
				typ.Kind(),
			)
		}
		schema.mutationType = typ
	}
	if cfg.SubscriptionType != nil {
		typ, err := newTypeRef(schema, cfg.SubscriptionType)
		if err != nil {
			return nil, fmt.Errorf("invalid config: subscriptino type: %s", err)
		}
		if _, ok := schema.typesMap[typ.name]; !ok && typ.name != "" {
			return nil, fmt.Errorf("invalid config: subscription type references non-existing type \"%s\"", typ.name)
		}
		if typ.Kind() != TypeKindObject {
			return nil, fmt.Errorf(
				"invalid config: subscription type must reference an OBJECT type, found %s type",
				typ.Kind(),
			)
		}
		schema.subscriptionType = typ
	}

	for _, t := range schema.types {
		// index object types by interfaces
		if t.Kind() == TypeKindObject {
			for _, i := range t.Interfaces() {
				schema.typesByInterface[i.Name()] = append(schema.typesByInterface[i.Name()], t)
			}
		}

		// check possible types in union types
		if t.Kind() == TypeKindUnion {
			for _, p := range t.PossibleTypes() {
				if p.Kind() == "" {
					return nil, fmt.Errorf(
						"invalid config: in UNION type \"%s\": possible type references non-existing type \"%s\"",
						t.Name(),
						p.Name(),
					)
				} else if p.Kind() != TypeKindObject {
					return nil, fmt.Errorf(
						"invalid config: in UNION type \"%s\": all possible types must reference OBJECT types, found %s type \"%s\"",
						t.Name(),
						p.Kind(),
						p.Name(),
					)
				}
			}
		}
	}

	/**
	TODO: add some validation
		- check all type refs reference existing types
		- check all input values are of input types
		- check all interfaces types are of interface kind
	*/

	return schema, nil
}

type schema struct {
	queryType        Type
	mutationType     Type
	subscriptionType Type
	types            []Type
	typesMap         map[string]Type
	directives       []Directive
	directivesMap    map[string]Directive
	typesByInterface map[string][]Type
}

func (s *schema) QueryType() Type                 { return s.queryType }
func (s *schema) MutationType() Type              { return s.mutationType }
func (s *schema) SubscriptionType() Type          { return s.subscriptionType }
func (s *schema) Types() []Type                   { return s.types }
func (s *schema) Type(name string) Type           { return s.typesMap[name] }
func (s *schema) Directives() []Directive         { return s.directives }
func (s *schema) Directive(name string) Directive { return s.directivesMap[name] }
