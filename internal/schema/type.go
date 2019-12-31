package schema

import (
	"errors"
	"fmt"
)

type Type interface {
	Kind() TypeKind
	Name() string
	Description() string

	// OBJECT and INTERFACE only
	Fields() []Field
	Field(string) Field

	// OBJECT ONLY
	Interfaces() []Type

	// INTERFACE and UNION only
	PossibleTypes() []Type

	// ENUM only
	EnumValues() []EnumValue

	// INPUT_OBJECT only
	InputFields() []InputValue
	InputField(string) InputValue

	// NON_NULL and LIST only
	OfType() Type
}

type TypeKind string

const (
	TypeKindScalar      TypeKind = "SCALAR"
	TypeKindObject               = "OBJECT"
	TypeKindInterface            = "INTERFACE"
	TypeKindUnion                = "UNION"
	TypeKindEnum                 = "ENUM"
	TypeKindInputObject          = "INPUT_OBJECT"
	TypeKindList                 = "LIST"
	TypeKindNonNull              = "NON_NULL"
)

var _ Type = (*typ)(nil)
var _ Type = (*typRef)(nil)

type TypeConfig struct {
	Kind          TypeKind            `json:"kind"`
	Name          string              `json:"name"`
	Description   string              `json:"description,omitempty"`
	Fields        []*FieldConfig      `json:"fields,omitempty"`
	Interfaces    []*TypeRefConfig    `json:"interfaces,omitempty"`
	PossibleTypes []*TypeRefConfig    `json:"possibleTypes,omitempty"`
	EnumValues    []*EnumValueConfig  `json:"enumValues,omitempty"`
	InputFields   []*InputValueConfig `json:"inputFields,omitempty"`
}

type TypeRefConfig struct {
	Kind   TypeKind       `json:"kind"`
	Name   string         `json:"name"`
	OfType *TypeRefConfig `json:"ofType,omitempty"`
}

func newType(schema *schema, cfg *TypeConfig) (Type, error) {
	if schema == nil {
		return nil, errors.New("missing schema")
	}
	if cfg == nil {
		return nil, errors.New("missing config")
	}
	if cfg.Name == "" {
		return nil, errors.New("missing name")
	}
	t := &typ{
		schema:         schema,
		kind:           cfg.Kind,
		name:           cfg.Name,
		description:    cfg.Description,
		fieldsMap:      map[string]Field{},
		inputFieldsMap: map[string]InputValue{},
	}

	switch cfg.Kind {
	case TypeKindScalar:
	case TypeKindEnum:
		// build enum values
		evIdx := map[string]bool{}
		for _, evCfg := range cfg.EnumValues {
			ev, err := newEnumValue(evCfg)
			if err != nil {
				return nil, fmt.Errorf(
					"in type \"%s\": %s",
					cfg.Name,
					err,
				)
			}
			if evIdx[ev.name] {
				return nil, fmt.Errorf(
					"in type \"%s\": enum value \"%s\" declared more than once",
					t.name,
					ev.name,
				)
			}
			evIdx[ev.name] = true
			t.enumValues = append(t.enumValues, ev)
		}
	case TypeKindObject, TypeKindInterface:
		// build fields and interfaces
		for _, fieldCfg := range cfg.Fields {
			f, err := newField(schema, fieldCfg)
			if err != nil {
				return nil, fmt.Errorf("in type \"%s\": %s", t.name, err)
			}
			if _, ok := t.fieldsMap[f.name]; ok {
				return nil, fmt.Errorf(
					"in type \"%s\": field \"%s\" declared more than once",
					t.name,
					f.name,
				)
			}
			t.fieldsMap[f.name] = f
			t.fields = append(t.fields, f)
		}
		if t.kind == TypeKindObject {
			idx := map[string]bool{}
			for _, ifaceCfg := range cfg.Interfaces {
				i, err := newTypeRef(schema, ifaceCfg)
				if err != nil {
					return nil, fmt.Errorf("in type \"%s\": %s", t.name, err)
				}
				if i.name == "" {
					return nil, fmt.Errorf(
						"in type \"%s\": interface type name can not be empty",
						t.name,
					)
				}
				if idx[t.name] {
					continue
				}
				t.interfaces = append(t.interfaces, i)
			}
		}
	case TypeKindInputObject:
		// build input fields
		for _, iCfg := range cfg.InputFields {
			i, err := newInputValue(schema, iCfg)
			if err != nil {
				return nil, fmt.Errorf("in type \"%s\": %s", t.name, err)
			}
			if _, ok := t.inputFieldsMap[i.name]; ok {
				return nil, fmt.Errorf(
					"in type \"%s\": input field \"%s\" declared more than once",
					t.name,
					i.name,
				)
			}
			t.inputFieldsMap[i.name] = i
			t.inputFields = append(t.inputFields, i)
		}
	case TypeKindUnion:
		// build possible types
		idx := map[string]bool{}
		for _, ptCfg := range cfg.PossibleTypes {
			pt, err := newTypeRef(schema, ptCfg)
			if err != nil {
				return nil, fmt.Errorf("in UNION type \"%s\": %s", t.name, err)
			}
			if pt.kind == TypeKindNonNull || pt.kind == TypeKindList {
				return nil, fmt.Errorf(
					"in UNION type \"%s\": all possible types must reference OBJECT types, found %s type",
					t.name,
					pt.kind,
				)
			}
			if idx[pt.name] {
				continue
			}
			t.possibleTypes = append(t.possibleTypes, pt)
		}
	case TypeKindList:
	case TypeKindNonNull:
		return nil, fmt.Errorf(
			"type kind \"%s\" can only be used in TypeRefConfig",
			cfg.Kind,
		)
	default:
		return nil, fmt.Errorf("unknown type kind \"%s\"", cfg.Kind)
	}
	return t, nil
}

type typ struct {
	schema         *schema
	kind           TypeKind
	name           string
	description    string
	fields         []Field
	fieldsMap      map[string]Field
	interfaces     []Type
	possibleTypes  []Type
	enumValues     []EnumValue
	inputFields    []InputValue
	inputFieldsMap map[string]InputValue
}

func (t *typ) Kind() TypeKind          { return t.kind }
func (t *typ) Name() string            { return t.name }
func (t *typ) Description() string     { return t.description }
func (t *typ) Fields() []Field         { return t.fields }
func (t *typ) Field(name string) Field { return t.fieldsMap[name] }
func (t *typ) Interfaces() []Type      { return t.interfaces }
func (t *typ) PossibleTypes() []Type {
	if t.kind == TypeKindInterface {
		return t.schema.typesByInterface[t.Name()]
	}
	return t.possibleTypes
}
func (t *typ) EnumValues() []EnumValue           { return t.enumValues }
func (t *typ) InputFields() []InputValue         { return t.inputFields }
func (t *typ) InputField(name string) InputValue { return t.inputFieldsMap[name] }
func (t *typ) OfType() Type                      { return nil }

func newTypeRef(schema *schema, cfg *TypeRefConfig) (*typRef, error) {
	if schema == nil {
		return nil, errors.New("missing schema")
	}
	if cfg == nil {
		return nil, errors.New("missing config")
	}

	ref := &typRef{
		schema: schema,
		kind:   cfg.Kind,
	}

	switch cfg.Kind {
	case TypeKindList, TypeKindNonNull:
		ofType, err := newTypeRef(schema, cfg.OfType)
		if err != nil {
			return nil, fmt.Errorf("in type ref: %s of type: %s", cfg.Kind, err)
		}
		if ref.kind == TypeKindNonNull && ofType.kind == TypeKindNonNull {
			return nil, errors.New("in type ref: NON_NULL of type: can not be NON_NULL type")
		}
		ref.ofType = ofType
	default:
		ref.name = cfg.Name
		if ref.name == "" {
			return nil, errors.New("in type ref: missing name")
		}
	}

	return ref, nil
}

type typRef struct {
	schema *schema
	kind   TypeKind
	name   string
	ofType Type
}

func (t *typRef) typ() Type { return t.schema.typesMap[t.name] }
func (t *typRef) Kind() TypeKind {
	// we only trust NON_NULL and LIST ref kinds
	if t.kind == TypeKindNonNull || t.kind == TypeKindList {
		return t.kind
	}
	// otherwise we fallback to the type kind from schema
	st := t.typ()
	if st == nil {
		return ""
	}
	return st.Kind()
}
func (t *typRef) Name() string { return t.name }
func (t *typRef) Description() string {
	st := t.typ()
	if st == nil {
		return ""
	}
	return st.Description()
}
func (t *typRef) Fields() []Field {
	st := t.typ()
	if st == nil {
		return nil
	}
	return st.Fields()
}
func (t *typRef) Field(name string) Field {
	st := t.typ()
	if st == nil {
		return nil
	}
	return st.Field(name)
}
func (t *typRef) Interfaces() []Type {
	st := t.typ()
	if st == nil {
		return nil
	}
	return st.Interfaces()
}
func (t *typRef) PossibleTypes() []Type {
	st := t.typ()
	if st == nil {
		return nil
	}
	return st.PossibleTypes()
}
func (t *typRef) EnumValues() []EnumValue {
	st := t.typ()
	if st == nil {
		return nil
	}
	return st.EnumValues()
}
func (t *typRef) InputFields() []InputValue {
	st := t.typ()
	if st == nil {
		return nil
	}
	return st.InputFields()
}
func (t *typRef) InputField(name string) InputValue {
	st := t.typ()
	if st == nil {
		return nil
	}
	return st.InputField(name)
}
func (t *typRef) OfType() Type {
	return t.ofType
}
