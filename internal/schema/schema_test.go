package schema

import (
	"testing"
)

func TestNewSchema_test_schema(t *testing.T) {
	_, err := NewSchema(newTestSchemaConfig())
	if err != nil {
		t.Errorf("NewSchema() returned error: %s", err)
	}
}

func TestNewSchema_possibleTypes_on_interface(t *testing.T) {
	cfg := newTestSchemaConfig()
	for _, t := range cfg.Types {
		if t.Name == "Character" {
			t.PossibleTypes = []*TypeRefConfig{&TypeRefConfig{Name: "SearchResult"}}
		}
	}
	schema, err := NewSchema(cfg)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	actual := []string{}
	for _, posType := range schema.Type("Character").PossibleTypes() {
		actual = append(actual, posType.Name())
	}
	checkSameElements(t, actual, []string{"Human", "Droid"})
}

func TestNewSchema_test_schema_without_mutation_type(t *testing.T) {
	cfg := newTestSchemaConfig()
	cfg.MutationType = nil
	_, err := NewSchema(newTestSchemaConfig())
	if err != nil {
		t.Errorf("NewSchema() returned error: %s", err)
	}
}

func TestNewSchema_test_schema_without_subscription_type(t *testing.T) {
	cfg := newTestSchemaConfig()
	cfg.SubscriptionType = nil
	_, err := NewSchema(newTestSchemaConfig())
	if err != nil {
		t.Errorf("NewSchema() returned error: %s", err)
	}
}

func TestNewSchema_root_type_config_errors(t *testing.T) {
	type cfgFn func(cfg *SchemaConfig)
	tests := []struct {
		name   string
		upCfg  cfgFn
		errMsg string
	}{
		{
			name: "config without a query type",
			upCfg: func(cfg *SchemaConfig) {
				cfg.QueryType = nil
			},
			errMsg: "invalid config: no query type specified",
		},
		{
			name: "config with a query type of kind NON_NULL",
			upCfg: func(cfg *SchemaConfig) {
				cfg.QueryType = &TypeRefConfig{Kind: TypeKindNonNull, OfType: cfg.QueryType}
			},
			errMsg: "invalid config: query type must reference an OBJECT type, found NON_NULL type",
		},
		{
			name: "config with a query type of kind INTERFACE",
			upCfg: func(cfg *SchemaConfig) {
				cfg.QueryType = &TypeRefConfig{Name: "Character"}
			},
			errMsg: "invalid config: query type must reference an OBJECT type, found INTERFACE type",
		},
		{
			name: "config with a query type that does not exist",
			upCfg: func(cfg *SchemaConfig) {
				cfg.QueryType = &TypeRefConfig{Name: "Plop"}
			},
			errMsg: "invalid config: query type references non-existing type \"Plop\"",
		},
		{
			name: "config with a mutation type of kind LIST",
			upCfg: func(cfg *SchemaConfig) {
				cfg.MutationType = &TypeRefConfig{Kind: TypeKindList, OfType: cfg.MutationType}
			},
			errMsg: "invalid config: mutation type must reference an OBJECT type, found LIST type",
		},
		{
			name: "config with a mutation type of kind INTERFACE",
			upCfg: func(cfg *SchemaConfig) {
				cfg.MutationType = &TypeRefConfig{Name: "Character"}
			},
			errMsg: "invalid config: mutation type must reference an OBJECT type, found INTERFACE type",
		},
		{
			name: "config with a mutation type that does not exist",
			upCfg: func(cfg *SchemaConfig) {
				cfg.MutationType = &TypeRefConfig{Name: "Plop"}
			},
			errMsg: "invalid config: mutation type references non-existing type \"Plop\"",
		},
		{
			name: "config with a subscription type of kind NON_NULL",
			upCfg: func(cfg *SchemaConfig) {
				cfg.SubscriptionType = &TypeRefConfig{Kind: TypeKindNonNull, OfType: cfg.SubscriptionType}
			},
			errMsg: "invalid config: subscription type must reference an OBJECT type, found NON_NULL type",
		},
		{
			name: "config with a subscription type of kind INTERFACE",
			upCfg: func(cfg *SchemaConfig) {
				cfg.SubscriptionType = &TypeRefConfig{Name: "Character"}
			},
			errMsg: "invalid config: subscription type must reference an OBJECT type, found INTERFACE type",
		},
		{
			name: "config with a subscription type that does not exist",
			upCfg: func(cfg *SchemaConfig) {
				cfg.SubscriptionType = &TypeRefConfig{Name: "Plop"}
			},
			errMsg: "invalid config: subscription type references non-existing type \"Plop\"",
		},
		{
			name: "config with a union that references a named type with non-OBJECT kind",
			upCfg: func(cfg *SchemaConfig) {
				for _, t := range cfg.Types {
					if t.Name == "SearchResult" {
						t.PossibleTypes = append(t.PossibleTypes, &TypeRefConfig{Name: "Character"})
					}
				}
			},
			errMsg: "invalid config: in UNION type \"SearchResult\": all possible types must reference OBJECT types, found INTERFACE type \"Character\"",
		},
		{
			name: "config with a union that references a non-named type",
			upCfg: func(cfg *SchemaConfig) {
				for _, t := range cfg.Types {
					if t.Name == "SearchResult" {
						t.PossibleTypes = append(t.PossibleTypes, &TypeRefConfig{Kind: TypeKindNonNull, OfType: &TypeRefConfig{Name: "Review"}})
					}
				}
			},
			errMsg: "invalid config: in UNION type \"SearchResult\": all possible types must reference OBJECT types, found NON_NULL type",
		},
		{
			name: "config with a union that references a non-existing named type",
			upCfg: func(cfg *SchemaConfig) {
				for _, t := range cfg.Types {
					if t.Name == "SearchResult" {
						t.PossibleTypes = append(t.PossibleTypes, &TypeRefConfig{Name: "Plop"})
					}
				}
			},
			errMsg: "invalid config: in UNION type \"SearchResult\": possible type references non-existing type \"Plop\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newTestSchemaConfig()
			tt.upCfg(cfg)
			got, err := NewSchema(cfg)
			checkError(t, err, tt.errMsg)
			if got != nil {
				t.Errorf("NewSchema() was not expected to return a Schema, but it did")
			}
		})
	}
}

func checkError(t *testing.T, actual error, expected string) {
	if actual == nil {
		t.Error("Error expected, none returned.")
		return
	}
	if actual.Error() != expected {
		t.Errorf("Error actual = %v, and Expected = %v.", actual, expected)
	}
}

func checkSameElements(t *testing.T, actual, expected []string) {
	actualLen := len(actual)
	expectedLen := len(expected)
	if actualLen != expectedLen {
		t.Errorf("Len of the lists don't match , len actual %v, len expected %v", actualLen, expectedLen)
	}
	visited := make([]bool, expectedLen)
	for i := 0; i < actualLen; i++ {
		found := false
		element := actual[i]
		for j := 0; j < expectedLen; j++ {
			if visited[j] {
				continue
			}
			if element == expected[j] {
				visited[j] = true
				found = true
				break
			}
		}
		if !found {
			t.Errorf("element %s appears more times in %s than in %s", element, actual, expected)
		}
	}
}

func newTestSchemaConfig() *SchemaConfig {
	return &SchemaConfig{
		QueryType:        &TypeRefConfig{Name: "Query"},
		MutationType:     &TypeRefConfig{Name: "Mutation"},
		SubscriptionType: &TypeRefConfig{Name: "Subscription"},
		Types: []*TypeConfig{
			&TypeConfig{
				Kind: TypeKindObject,
				Name: "Query",
				Fields: []*FieldConfig{
					&FieldConfig{
						Name: "hero",
						Args: []*InputValueConfig{
							&InputValueConfig{
								Name: "episode",
								Type: &TypeRefConfig{Name: "Episode"},
							},
						},
						Type: &TypeRefConfig{Name: "Character"},
					},
					&FieldConfig{
						Name: "droid",
						Args: []*InputValueConfig{
							&InputValueConfig{
								Name: "id",
								Type: &TypeRefConfig{
									Kind:   TypeKindNonNull,
									OfType: &TypeRefConfig{Name: "ID"},
								},
							},
						},
						Type: &TypeRefConfig{Name: "Droid"},
					},
					&FieldConfig{
						Name: "search",
						Args: []*InputValueConfig{
							&InputValueConfig{
								Name: "text",
								Type: &TypeRefConfig{
									Kind:   TypeKindNonNull,
									OfType: &TypeRefConfig{Name: "String"},
								},
							},
						},
						Type: &TypeRefConfig{
							Kind:   TypeKindList,
							OfType: &TypeRefConfig{Name: "Droid"},
						},
					},
				},
			},
			&TypeConfig{
				Kind: TypeKindObject,
				Name: "Mutation",
				Fields: []*FieldConfig{
					&FieldConfig{
						Name: "createReview",
						Args: []*InputValueConfig{
							&InputValueConfig{
								Name: "episode",
								Type: &TypeRefConfig{
									Kind:   TypeKindNonNull,
									OfType: &TypeRefConfig{Name: "Episode"},
								},
							},
							&InputValueConfig{
								Name: "review",
								Type: &TypeRefConfig{
									Kind:   TypeKindNonNull,
									OfType: &TypeRefConfig{Name: "ReviewInput"},
								},
							},
						},
						Type: &TypeRefConfig{Name: "Review"},
					},
				},
			},
			&TypeConfig{
				Kind: TypeKindObject,
				Name: "Subscription",
				Fields: []*FieldConfig{
					&FieldConfig{
						Name: "newReview",
						Type: &TypeRefConfig{Name: "Review"},
					},
				},
			},
			&TypeConfig{
				Kind: TypeKindInterface,
				Name: "Character",
				Fields: []*FieldConfig{
					&FieldConfig{
						Name: "id",
						Type: &TypeRefConfig{
							Kind:   TypeKindNonNull,
							OfType: &TypeRefConfig{Name: "ID"},
						},
					},
					&FieldConfig{
						Name: "name",
						Type: &TypeRefConfig{
							Kind:   TypeKindNonNull,
							OfType: &TypeRefConfig{Name: "String"},
						},
					},
					&FieldConfig{
						Name: "friends",
						Type: &TypeRefConfig{
							Kind:   TypeKindList,
							OfType: &TypeRefConfig{Name: "Episode"},
						},
					},
					&FieldConfig{
						Name: "appearsIn",
						Type: &TypeRefConfig{
							Kind: TypeKindNonNull,
							OfType: &TypeRefConfig{
								Kind: TypeKindList,
								OfType: &TypeRefConfig{
									Kind:   TypeKindNonNull,
									OfType: &TypeRefConfig{Name: "Episode"},
								},
							},
						},
					},
				},
			},
			&TypeConfig{
				Kind: TypeKindObject,
				Name: "Human",
				Interfaces: []*TypeRefConfig{
					&TypeRefConfig{Name: "Character"},
				},
				Fields: []*FieldConfig{
					&FieldConfig{
						Name: "id",
						Type: &TypeRefConfig{
							Kind:   TypeKindNonNull,
							OfType: &TypeRefConfig{Name: "ID"},
						},
					},
					&FieldConfig{
						Name: "name",
						Type: &TypeRefConfig{
							Kind:   TypeKindNonNull,
							OfType: &TypeRefConfig{Name: "String"},
						},
					},
					&FieldConfig{
						Name: "friends",
						Type: &TypeRefConfig{
							Kind:   TypeKindList,
							OfType: &TypeRefConfig{Name: "Episode"},
						},
					},
					&FieldConfig{
						Name: "appearsIn",
						Type: &TypeRefConfig{
							Kind: TypeKindNonNull,
							OfType: &TypeRefConfig{
								Kind: TypeKindList,
								OfType: &TypeRefConfig{
									Kind:   TypeKindNonNull,
									OfType: &TypeRefConfig{Name: "Episode"},
								},
							},
						},
					},
					&FieldConfig{
						Name: "starships",
						Type: &TypeRefConfig{
							Kind:   TypeKindList,
							OfType: &TypeRefConfig{Name: "Starship"},
						},
					},
					&FieldConfig{
						Name: "totalCredits",
						Type: &TypeRefConfig{Name: "Int"},
					},
				},
			},
			&TypeConfig{
				Kind: TypeKindObject,
				Name: "Droid",
				Interfaces: []*TypeRefConfig{
					&TypeRefConfig{Name: "Character"},
				},
				Fields: []*FieldConfig{
					&FieldConfig{
						Name: "id",
						Type: &TypeRefConfig{
							Kind:   TypeKindNonNull,
							OfType: &TypeRefConfig{Name: "ID"},
						},
					},
					&FieldConfig{
						Name: "name",
						Type: &TypeRefConfig{
							Kind:   TypeKindNonNull,
							OfType: &TypeRefConfig{Name: "String"},
						},
					},
					&FieldConfig{
						Name: "friends",
						Type: &TypeRefConfig{
							Kind:   TypeKindList,
							OfType: &TypeRefConfig{Name: "Episode"},
						},
					},
					&FieldConfig{
						Name: "appearsIn",
						Type: &TypeRefConfig{
							Kind: TypeKindNonNull,
							OfType: &TypeRefConfig{
								Kind: TypeKindList,
								OfType: &TypeRefConfig{
									Kind:   TypeKindNonNull,
									OfType: &TypeRefConfig{Name: "Episode"},
								},
							},
						},
					},
					&FieldConfig{
						Name: "primaryFunction",
						Type: &TypeRefConfig{Name: "String"},
					},
				},
			},
			&TypeConfig{
				Kind: TypeKindObject,
				Name: "Starship",
				Fields: []*FieldConfig{
					&FieldConfig{
						Name: "id",
						Type: &TypeRefConfig{
							Kind:   TypeKindNonNull,
							OfType: &TypeRefConfig{Name: "ID"},
						},
					},
					&FieldConfig{
						Name: "name",
						Type: &TypeRefConfig{
							Kind:   TypeKindNonNull,
							OfType: &TypeRefConfig{Name: "String"},
						},
					},
				},
			},
			&TypeConfig{
				Kind: TypeKindUnion,
				Name: "SearchResult",
				PossibleTypes: []*TypeRefConfig{
					&TypeRefConfig{Name: "Human"},
					&TypeRefConfig{Name: "Droid"},
					&TypeRefConfig{Name: "Starship"},
				},
			},
			&TypeConfig{
				Kind: TypeKindEnum,
				Name: "Episode",
				EnumValues: []*EnumValueConfig{
					&EnumValueConfig{Name: "NEWHOPE"},
					&EnumValueConfig{Name: "EMPIRE"},
					&EnumValueConfig{Name: "JEDI"},
				},
			},
			&TypeConfig{
				Kind: TypeKindObject,
				Name: "Review",
				Fields: []*FieldConfig{
					&FieldConfig{
						Name: "stars",
						Type: &TypeRefConfig{
							Kind:   TypeKindNonNull,
							OfType: &TypeRefConfig{Name: "Int"},
						},
					},
					&FieldConfig{
						Name: "commentary",
						Type: &TypeRefConfig{
							Kind:   TypeKindNonNull,
							OfType: &TypeRefConfig{Name: "String"},
						},
					},
				},
			},
			&TypeConfig{
				Kind: TypeKindInputObject,
				Name: "ReviewInput",
				InputFields: []*InputValueConfig{
					&InputValueConfig{
						Name: "stars",
						Type: &TypeRefConfig{
							Kind:   TypeKindNonNull,
							OfType: &TypeRefConfig{Name: "Int"},
						},
					},
					&InputValueConfig{
						Name: "commentary",
						Type: &TypeRefConfig{
							Kind:   TypeKindNonNull,
							OfType: &TypeRefConfig{Name: "String"},
						},
					},
				},
			},
			&TypeConfig{
				Kind: TypeKindScalar,
				Name: "ID",
			},
			&TypeConfig{
				Kind: TypeKindScalar,
				Name: "Int",
			},
			&TypeConfig{
				Kind: TypeKindScalar,
				Name: "Float",
			},
			&TypeConfig{
				Kind: TypeKindScalar,
				Name: "String",
			},
			&TypeConfig{
				Kind: TypeKindScalar,
				Name: "Boolean",
			},
		},
	}
}
