package redisearch

import (
	"fmt"
)

func FtSearch(idx string, query Query) []interface{} {
	clause := []interface{}{
		"FT.SEARCH",
		idx,
	}
	clause = append(clause, query.Args()...)
	return clause

}

func FtDropIndex(idxes ...string) []interface{} {
	clause := []interface{}{
		"FT.DROPINDEX",
	}
	for _, idx := range idxes {
		clause = append(clause, idx)
	}
	return clause
}

type FtCreateOn string

const (
	FTCREATE_ON_HASH FtCreateOn = "HASH"
	FTCREATE_ON_JSON FtCreateOn = "JSON"
)

func FtCreate(idx string, on FtCreateOn, prefixes []string, schemas []Schema) []interface{} {
	clause := []interface{}{
		"FT.CREATE",
		idx,
		"ON",
		string(on),
	}
	prefix_cnt := len(prefixes)
	prefix_clause := []interface{}{"PREFIX", fmt.Sprintf("%d", prefix_cnt)}
	for _, prefix := range prefixes {
		prefix_clause = append(prefix_clause, prefix)
	}
	clause = append(clause, prefix_clause...)
	schema_clause := []interface{}{"SCHEMA"}
	for _, schema := range schemas {
		field := schema.field
		if on == FTCREATE_ON_JSON {
			field = "$." + field
		}
		schema_clause = append(schema_clause, field)
		if schema.alias != "" {
			schema_clause = append(schema_clause, "AS", schema.alias)
		}
		schema_clause = append(schema_clause, schema.kind)
	}
	clause = append(clause, schema_clause...)
	return clause
}

func FtConfigSet(option string, value string) []interface{} {
	return []interface{}{"FT.CONFIG", "SET", option, value}
}
