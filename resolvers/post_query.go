package resolvers

import (
	"context"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/lambdacollective/cobbles-api/server"
	sq "gopkg.in/Masterminds/squirrel.v1"
	squirrel "gopkg.in/Masterminds/squirrel.v1"
)

func (r *Resolver) scanPost(row scannable) (*server.Post, error) {
	var p server.Post
	var result struct {
		createdAt  pgtype.Timestamptz
		processing pgtype.Bool
	}
	err := row.Scan(
		&p.ID,
		&p.UserID,
		&p.NeighborhoodID,
		&p.Kind,
		&p.Title,
		&p.Description,
		&p.Poster,
		&p.UploadedMediaURL,
		&p.Media,
		&p.Preview,
		&p.Tags,
		&p.ViewTimes,
		&result.processing,
		&result.createdAt,
	)
	if err != nil {
		return nil, err
	}

	p.Processing = result.processing.Bool
	p.CreatedAt = result.createdAt.Time
	return &p, nil
}

func (r *Resolver) getPost(id int64) (*server.Post, bool, error) {
	sql, args, err := newSelectBuilder("id", "user_id", "neighborhood_id",
		"kind", "title", "description", "poster", "uploaded_media_url", "media", "preview",
		"tags", "view_times", "processing", "created_at").
		From("posts").
		Where(sq.Eq{"id": id}).
		ToSql()

	row := r.server.ConnPool.QueryRow(sql, args...)
	post, err := r.scanPost(row)
	if err == pgx.ErrNoRows {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, err
	}

	return post, true, nil
}

type resolvePostsInput struct {
	// (optional) Limit field to a user, eg currentUsers { posts() }
	ByUserID int64
	Tags     *[]string

	PageToken *string
	Limit     int32
}

// TODO: actual isolated models code?
func resolvePosts(ctx context.Context, s *server.Server, in resolvePostsInput) ([]*PostResolver, *string, error) {
	// TODO(tejasmanohar): sq.Where neighborhood is user's neighborhood by default
	sqlStmt := newSelectBuilder(
		"p.id",
		"p.user_id",
		"p.neighborhood_id",
		"p.title",
		"p.description",
		"p.kind",
		"p.tags",
		"p.poster",
		"p.media",
		"p.preview",
		"p.uploaded_media_url",
		"p.processing",
		"coalesce(p.likes, 0)",
		"p.comment_count",
		"p.view_times",
		"p.created_at",
		"p.updated_at",
		"n.id",
		"n.name",
		"n.slug").
		From("posts p").
		Where("p.removed is false").
		Where("p.processing is false").
		Join("neighborhoods n on n.id = p.neighborhood_id").
		OrderBy("created_at desc").
		Limit(uint64(in.Limit + 1))

	afterID, err := DecodeAfterIDCursor(in.PageToken)
	if err != nil {
		return nil, nil, err
	}

	if afterID > 0 {
		// Less than because we're paginating backwards
		sqlStmt = sqlStmt.Where(squirrel.Lt{"p.id": afterID})
	}

	if in.ByUserID > 0 {
		sqlStmt = sqlStmt.Where(squirrel.Eq{"user_id": in.ByUserID})
	}

	if in.Tags != nil {
		sqlStmt = sqlStmt.Where("tags && ?", in.Tags)
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

	var postResolvers []*PostResolver
	var lastID int64
	var i int32
	for rows.Next() {
		if i == in.Limit {
			lastID = postResolvers[len(postResolvers)-1].post.ID
			break
		}

		var neighborhood server.Neighborhood
		var post server.Post
		var result struct {
			postKind  string
			createdAt pgtype.Timestamptz
			updatedAt pgtype.Timestamptz
		}
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.NeighborhoodID,
			&post.Title,
			&post.Description,
			&result.postKind,
			&post.Tags,
			&post.Poster,
			&post.Media,
			&post.Preview,
			&post.UploadedMediaURL,
			&post.Processing,
			&post.Likes,
			&post.CommentCount,
			&post.ViewTimes,
			&result.createdAt,
			&result.updatedAt,
			&neighborhood.ID,
			&neighborhood.Name,
			&neighborhood.Slug,
		)
		if err != nil {
			return nil, nil, err
		}

		post.Kind = server.PostKind(strings.ToUpper(result.postKind))
		post.CreatedAt = result.createdAt.Time
		post.UpdatedAt = result.updatedAt.Time

		postResolvers = append(postResolvers, &PostResolver{
			server:       s,
			post:         &post,
			neighborhood: &neighborhood,
		})
		i++
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return postResolvers, EncodeAfterIDCursor(lastID), nil
}

func postByID(s *server.Server, postID int64) (*server.Post, error) {
	var p server.Post
	var result struct {
		createdAt  pgtype.Timestamptz
		processing pgtype.Bool
	}
	err := s.ConnPool.QueryRow(`
		select
			id,
			user_id,
			neighborhood_id,
			kind,
			title,
			description,
			poster,
			uploaded_media_url,
			media,
			preview,
			tags,
			processing,
			created_at,
			likes
		from posts
		where id = $1
	`, postID).Scan(
		&p.ID,
		&p.UserID,
		&p.NeighborhoodID,
		&p.Kind,
		&p.Title,
		&p.Description,
		&p.Poster,
		&p.UploadedMediaURL,
		&p.Media,
		&p.Preview,
		&p.Tags,
		&result.processing,
		&result.createdAt,
		&p.Likes,
	)
	if err != nil {
		return nil, err
	}

	p.Processing = result.processing.Bool
	p.CreatedAt = result.createdAt.Time

	return &p, nil
}

// HasCurrentUserLikedPost
func (r *Resolver) HasCurrentUserLikedPost(ctx context.Context, args struct {
	ID int32 // Post ID
}) (bool, error) {
	userID, err := ctxUserID(ctx)
	if err != nil {
		return false, err
	}

	status, err := r.server.HasUserLikedPost(int32(userID), args.ID)

	return status, err

}
