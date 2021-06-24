package resolvers

import (
	"context"
)

// isLiked
func (r *Resolver) IsLiked(ctx context.Context, args struct {
	ID int32
}) (bool, error) {

	// for test
	currentUserID, err := ctxUserID(ctx)
	if err != nil {
		return false, err
	}

	_, err = r.server.GetLike(int32(currentUserID), int32(args.ID))

	// _, err := r.server.GetLike(int32(9), int32(args.ID))
	// for test endks

	if err != nil {
		return false, nil
	}

	return true, nil
}
