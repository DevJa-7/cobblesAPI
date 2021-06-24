package resolvers

import "context"

// GetFollowingByUserID  - Who am I following
func (r *Resolver) GetFollowingByUserID(ctx context.Context, args struct {
	ID int32
}) (*FollowersResult, error) {
	users, err := r.server.GetFollowingByUserID(args.ID)

	if err != nil {
		return nil, err
	}

	var userResolverArray []*UserResolver
	for _, u := range users {
		userResolverArray = append(userResolverArray, &UserResolver{
			server: r.server,
			user:   u,
		})
	}

	return &FollowersResult{followers: userResolverArray}, nil
}

// GetFollowersByUserID  - Who are my followers
func (r *Resolver) GetFollowersByUserID(ctx context.Context, args struct {
	ID int32
}) (*FollowersResult, error) {
	users, err := r.server.GetFollowersByUserID(args.ID)

	if err != nil {
		return nil, err
	}

	var userResolverArray []*UserResolver
	for _, u := range users {
		userResolverArray = append(userResolverArray, &UserResolver{
			server: r.server,
			user:   u,
		})
	}

	return &FollowersResult{followers: userResolverArray}, nil
}
