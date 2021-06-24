package resolvers

import "context"

// LikePost ...
func (r *Resolver) LikePost(ctx context.Context, args struct {
	ID int32 // Post ID
}) (bool, error) {
	// for test
	userID, err := ctxUserID(ctx)
	if err != nil {
		return false, err
	}
	_, err = r.server.Like(int32(userID), args.ID)
	// userID := 9
	// _, err := r.server.Like(int32(userID), args.ID)
	// for test end

	if err != nil {
		return false, err
	}

	return true, nil
}

// UnlikePost ...
func (r *Resolver) UnlikePost(ctx context.Context, args struct {
	ID int32 // Post ID
}) (bool, error) {
	userID, err := ctxUserID(ctx)
	if err != nil {
		return false, err
	}
	_, err = r.server.Unlike(int32(userID), args.ID)

	if err != nil {
		return false, err
	}

	return true, nil
}
