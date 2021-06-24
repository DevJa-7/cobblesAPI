package gqlschema

import (
	graphql "github.com/graph-gophers/graphql-go"
)

func MustParseSchema(resolver interface{}) *graphql.Schema {
	return graphql.MustParseSchema(typeDefs, resolver)
}
