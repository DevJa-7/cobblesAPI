package resolvers

import "context"

// CreatePostComment ...
func (r *Resolver) CreatePostComment(ctx context.Context, args struct {
	Input CreatePostCommentInput
}) (bool, error) {
	userID, err := ctxUserID(ctx)
	if err != nil {
		return false, err
	}
	_, err = r.server.CreatePostComment(
		args.Input.ParentCommentID,
		int32(userID),
		args.Input.PostID,
		args.Input.Comment,
	)

	if err != nil {
		return false, err
	}

	return true, nil
}

// RemovePostComment ...
func (r *Resolver) RemovePostComment(ctx context.Context, args struct {
	Input RemovePostCommentInput
}) (bool, error) {
	err := r.server.RemovePostComment(
		args.Input.CommentID,
		args.Input.PostID,
	)

	if err != nil {
		return false, err
	}

	return true, nil
}
