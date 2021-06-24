package resolvers

import "time"

func (r *Resolver) Hello() string {
	return time.Now().UTC().Format(time.RFC3339)
}
