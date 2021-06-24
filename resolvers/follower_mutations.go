package resolvers

import (
	"context"
)

// CreateFollower ...
func (r *Resolver) CreateFollower(ctx context.Context, args struct {
	ID int32 // Follower User ID
}) (*FollowerResolver, error) {
	userID, err := ctxUserID(ctx)
	if err != nil {
		return nil, err
	}
	follower, err := r.server.CreateFollower(int32(userID), args.ID)

	if err != nil {
		return nil, err
	}

	return &FollowerResolver{server: r.server, follower: follower}, nil
}

// Unfollow ...
func (r *Resolver) Unfollow(ctx context.Context, args struct {
	ID int32 // Follower User ID
}) (bool, error) {
	userID, err := ctxUserID(ctx)
	if err != nil {
		return false, err
	}

	status, err := r.server.Unfollow(int32(userID), args.ID)
	if err != nil {
		return false, err
	}

	return status, nil
}
