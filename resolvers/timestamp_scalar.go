package resolvers

import (
	"encoding/json"
	"fmt"
	"time"
)

// Timestamp is a custom GraphQL scalar for timestamps.
// graph-gophers/graphql-go graphql.Time / `scalar Time`, but it supports
// multiple formats and isn't very strict/typed, which is bad IMO.
type Timestamp struct {
	time.Time
}

// ImplementsGraphQLType selects "Timestamp"
func (t *Timestamp) ImplementsGraphQLType(name string) bool {
	return name == "Timestamp"
}

// UnmarshalGraphQL accepts an RFC 3339 string
func (t *Timestamp) UnmarshalGraphQL(input interface{}) error {
	str, ok := input.(string)
	if !ok {
		return fmt.Errorf("got %T, expected string", input)
	}

	var err error
	t.Time, err = time.Parse(time.RFC3339, str)
	return err
}

// MarshalJSON produces an RFC 3339Nano string
func (t *Timestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Time)
}
