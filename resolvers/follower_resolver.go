package resolvers

import (
	"strconv"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/lambdacollective/cobbles-api/server"
)

// Follower ...
type Follower struct {
	UserID         int32
	FollowerUserID int32
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// FollowersResult ...
type FollowersResult struct {
	followers []*UserResolver
}

// GetFollowersByUserID  - Who are my followers
func (r *FollowersResult) GetFollowersByUserID() []*UserResolver {
	return r.followers
}

// GetFollowingByUserID  - Who am I following
func (r *FollowersResult) GetFollowingByUserID() []*UserResolver {
	return r.followers
}

func (r *FollowersResult) Followers() []*UserResolver {
	return r.followers
}

// FollowerResolver ...
type FollowerResolver struct {
	server   *server.Server
	follower *server.Follower
}

func (r *FollowerResolver) ID() graphql.ID {
	return graphql.ID(strconv.FormatUint(uint64(r.follower.ID), 10))
}

func (r *FollowerResolver) UserID() int32 {
	return int32(r.follower.UserID)
}

func (r *FollowerResolver) FollowerUserID() int32 {
	return int32(r.follower.FollowerUserID)
}

func (r *FollowerResolver) CreatedAt() Timestamp {
	return Timestamp{r.follower.CreatedAt}
}

func (r *FollowerResolver) UpdatedAt() Timestamp {
	return Timestamp{r.follower.UpdatedAt}
}
