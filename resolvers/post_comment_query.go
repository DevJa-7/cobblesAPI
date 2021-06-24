package resolvers

import (
	"context"

	"github.com/jackc/pgx/pgtype"
	"github.com/lambdacollective/cobbles-api/server"
	squirrel "gopkg.in/Masterminds/squirrel.v1"
)

func (r *Resolver) GetCommentsByPostID(ctx context.Context, args struct {
	ID int32
}) (*PostCommentsResult, error) {
	return nil, nil
}

type resolvePostCommentsInput struct {
	// ByUserID int64
	PostID    int32
	PageToken *string
	Limit     int32
}

// resolve post comments
func resolvePostComments(s *server.Server, in resolvePostCommentsInput) ([]*PostCommentResolver, *string, error) {

	sqlStmt := newSelectBuilder(
		"pc.id",
		"pc.comment",
		"pc.parent_comment_id",
		"pc.user_id",
		"pc.post_id",
		"pc.created_at",
		"pc.updated_at").
		From("post_comments pc").
		OrderBy("created_at desc").
		Limit(uint64(in.Limit + 1))

	afterID, err := DecodeAfterIDCursor(in.PageToken)
	if err != nil {
		return nil, nil, err
	}

	if afterID > 0 {
		// Less than because we're paginating backwards
		sqlStmt = sqlStmt.Where(squirrel.Lt{"pc.id": afterID})
	}

	if in.PostID > 0 {
		sqlStmt = sqlStmt.Where(squirrel.Eq{"post_id": in.PostID})
	}

	sql, args, err := sqlStmt.ToSql()
	if err != nil {
		return nil, nil, err
	}

	rows, err := s.ConnPool.Query(sql, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var postCommentResolvers []*PostCommentResolver
	var lastID int64
	var i int32
	for rows.Next() {
		if i == in.Limit {
			lastID = postCommentResolvers[len(postCommentResolvers)-1].postComment.ID
			break
		}

		var postComment server.PostComment
		var result struct {
			postKind  string
			createdAt pgtype.Timestamptz
			updatedAt pgtype.Timestamptz
		}
		err := rows.Scan(
			&postComment.ID,
			&postComment.Comment,
			&postComment.ParentCommentID,
			&postComment.UserID,
			&postComment.PostID,
			&result.createdAt,
			&result.updatedAt,
		)
		if err != nil {
			return nil, nil, err
		}

		postComment.CreatedAt = result.createdAt.Time
		postComment.UpdatedAt = result.updatedAt.Time

		postCommentResolvers = append(postCommentResolvers, &PostCommentResolver{
			server:      s,
			postComment: &postComment,
		})
		i++
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return postCommentResolvers, EncodeAfterIDCursor(lastID), nil
}
