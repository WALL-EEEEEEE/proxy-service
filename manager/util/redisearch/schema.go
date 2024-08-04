package redisearch

type SchemaKind string

func (k SchemaKind) MarshalBinary() ([]byte, error) {
	return []byte(k), nil
}

func (k *SchemaKind) UnmarshalBinary(data []byte) error {
	*k = SchemaKind(string(data))
	return nil
}

const (
	SCHEMA_KIND_TAG     SchemaKind = "TAG"
	SCHEMA_KIND_TEXT    SchemaKind = "TEXT"
	SCHEMA_KIND_NUMERIC SchemaKind = "NUMERIC"
)

type SchemaOption func(*schemaOptions)

func AliasSchemaOption(alias string) SchemaOption {
	return func(options *schemaOptions) {
		options.alias = &alias
	}
}

type schemaOptions struct {
	alias *string
}

type Schema struct {
	field string
	alias string
	kind  SchemaKind
}

func NewSchema(field string, kind SchemaKind, opts ...SchemaOption) Schema {
	s := Schema{
		field: field,
		kind:  kind,
	}
	options := &schemaOptions{}
	for _, opt := range opts {
		opt(options)
	}
	if options.alias != nil {
		s.alias = *options.alias
	}
	return s
}
