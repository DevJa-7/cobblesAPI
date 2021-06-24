package resolvers

import (
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/lambdacollective/cobbles-api/server"
)

type NeighborhoodResolver struct {
	server *server.Server

	neighborhood *server.Neighborhood
}

func (r *NeighborhoodResolver) ID() graphql.ID {
	return graphql.ID(r.neighborhood.Slug)
}

func (r *NeighborhoodResolver) Slug() string {
	return r.neighborhood.Slug
}

func (r *NeighborhoodResolver) Name() string {
	return r.neighborhood.Name
}
