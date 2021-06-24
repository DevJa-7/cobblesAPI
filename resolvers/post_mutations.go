package resolvers

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/jackc/pgx"
	"github.com/lambdacollective/cobbles-api/server"
	"github.com/pkg/errors"

	graphql "github.com/graph-gophers/graphql-go"
)

func (r *Resolver) CreatePost(ctx context.Context, args struct {
	Input struct {
		Neighborhood string
		Title        string
		Description  *string

		Poster *string

		Kind          server.PostKind
		MediaKind     *string
		MediaURL      *string
		MediaMetadata *struct {
			Width  *int32 `json:"width"`
			Height *int32 `json:"height"`
		}

		Tags *[]string
	}
}) (*PostResolver, error) {
	currentUserID, err := ctxUserID(ctx)
	if err != nil {
		return nil, err
	}

	// inputNeighborhood := args.Input.Neighborhood
	inputTitle := args.Input.Title

	inputDescription := args.Input.Description
	inputPoster := args.Input.Poster
	inputTags := args.Input.Tags

	inputMediaKind := args.Input.MediaKind
	inputMediaURL := args.Input.MediaURL
	inputMediaMetadata := args.Input.MediaMetadata

	neighborhood, err := r.server.NeighborhoodBySlug("southie")
	if err != nil {
		return nil, err
	}

	// var mediaURL *string
	var uploadedMedia PostMedia
	var mediaURLBucket string

	if inputMediaURL != nil {
		mediaURL, err := url.Parse(*inputMediaURL)
		if err != nil {
			return nil, err
		}

		hostname := mediaURL.Hostname()
		if hostname != r.s3Hostname() {
			return nil, fmt.Errorf("invalid hostname: %s", hostname)
		}

		pathParts := strings.Split(mediaURL.Path, "/")
		if len(pathParts) > 1 {
			mediaURLBucket = strings.ToLower(pathParts[1])
		}

		uploadedMedia.URL = mediaURL.String()
	}

	if inputMediaMetadata != nil {
		if inputMediaMetadata.Height != nil {
			uploadedMedia.Height = inputMediaMetadata.Height
		}

		if inputMediaMetadata.Width != nil {
			uploadedMedia.Width = inputMediaMetadata.Width
		}
	}

	var postProcessing bool
	var postMedia *PostMedia
	var postPreview *PostMedia

	switch args.Input.Kind {
	case PostKindVideo:
		if mediaURLBucket != "videos" {
			return nil, errors.New("incorrect mediaURL for VIDEO")
		}

		postPreview = &PostMedia{Width: uploadedMedia.Width, Height: uploadedMedia.Height}
		postProcessing = true
	case PostKindImage:
		if mediaURLBucket != "images" {
			return nil, errors.New("incorrect mediaURL for IMAGE")
		}
		postMedia = &uploadedMedia
		postPreview = &uploadedMedia
	case PostKindText:
		if inputPoster == nil {
			return nil, errors.New("poster must be specified for type TEXT")
		}
	default:
		return nil, errors.New("post kind is invalid")
	}

	lowerInputKind := strings.ToLower(string(args.Input.Kind))
	row := r.server.ConnPool.QueryRow(`
		insert into posts (
			user_id,
			neighborhood_id,
			processing,
			kind,
			title,
			description,
			poster,
			media_kind,
			uploaded_media_url,
			uploaded_media,
			media,
			preview,
			tags,
			created_at,
			updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, now(), now())
		returning 
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
			view_times,
			processing,
			created_at
	`, currentUserID,
		neighborhood.ID,
		postProcessing,
		lowerInputKind,
		inputTitle,
		inputDescription,
		inputPoster,
		inputMediaKind,
		inputMediaURL,
		uploadedMedia,
		postMedia,
		postPreview,
		inputTags,
	)
	p, err := r.scanPost(row)
	if err != nil {
		return nil, err
	}

	// calculate the post count
	_, err = r.server.RecalculatePostCount(currentUserID)
	if err != nil {
		return nil, err
	}

	if args.Input.Kind == PostKindVideo {
		if err := r.server.ProcessPostMediaUpload(p.ID, uploadedMedia.URL); err != nil {
			log.Println(err)
		}
	}

	return &PostResolver{
		server: r.server,
		post:   p,
	}, nil
}

func (r *Resolver) UpdatePost(ctx context.Context, args struct {
	Input struct {
		ID          graphql.ID
		Poster      *string
		Title       *string
		Description *string
		Tags        *[]string
	}
}) (*PostResolver, error) {
	currentUserID, err := ctxUserID(ctx)
	if err != nil {
		return nil, err
	}

	inputPostID := args.Input.ID
	inputPoster := args.Input.Poster
	inputTitle := args.Input.Title
	inputDescription := args.Input.Description
	inputTags := args.Input.Tags

	row := r.server.ConnPool.QueryRow(`
		update posts
		set
			title = coalesce($3, title),
			description = coalesce($4, description),
			poster = coalesce($5, poster),
			tags = coalesce($6, tags),
			updated_at = now()
		where user_id = $1 and id = $2 and removed is false
		returning
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
			view_times,
			processing,
			created_at
	`, currentUserID, inputPostID, inputTitle, inputDescription, inputPoster, inputTags)

	p, err := r.scanPost(row)
	switch {
	case err == pgx.ErrNoRows:
		return nil, errors.New("post does not exist")
	case err != nil:
		log.Println(err)

		// TODO standardize internal server error responses
		return nil, errors.New("internal server error")
	}

	return &PostResolver{
		server: r.server,
		post:   p,
	}, nil
}

// RemovePostInput ...
type RemovePostInput struct {
	ID graphql.ID
}

func (r *Resolver) RemovePost(ctx context.Context, args struct {
	Input RemovePostInput
}) (*PostResolver, error) {
	currentUserID, err := ctxUserID(ctx)
	if err != nil {
		return nil, err
	}

	inputPostID := args.Input.ID
	row := r.server.ConnPool.QueryRow(`
		update posts
		set
			removed = true
		where
			user_id = $1
			and id = $2
			and removed is false
		returning
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
			view_times,
			processing,
			created_at
	`, currentUserID, inputPostID)

	p, err := r.scanPost(row)
	switch {
	case err == pgx.ErrNoRows:
		return nil, errors.New("post does not exist")
	case err != nil:
		log.Println(err)
		// TODO standardize internal server error responses
		return nil, errors.New("internal server error")
	}

	// calculate the post count
	_, err = r.server.RecalculatePostCount(currentUserID)
	if err != nil {
		return nil, err
	}

	return &PostResolver{
		server: r.server,
		post:   p,
	}, nil
}

// PostViewTimes : post view times are recored to the post table.
func (r *Resolver) PostViewTimes(ctx context.Context, args struct {
	PostID string
}) (*string, error) {
	var result struct {
		viewTimes int64
		userID    int64
	}

	currentUserID, err := ctxUserID(ctx)
	if err != nil {
		return nil, err
	}

	inputPostID, err := strconv.ParseInt(args.PostID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid post ID: %s", args.PostID)
	}

	err = r.server.ConnPool.QueryRow(`
		select
			user_id,
			view_times
		from posts
		where id = $1
	`, inputPostID).Scan(
		&result.userID,
		&result.viewTimes,
	)
	
	if err != nil {
		return nil, err
	}

	if currentUserID != result.userID {
		err = r.server.ConnPool.QueryRow(`
			update posts
			set view_times = $1
			where id = $2
			returning
				view_times
		`, (result.viewTimes + 1), inputPostID).Scan(
			&result.viewTimes,
		)
		if err != nil {
			return nil, err
		}

		res := strconv.FormatInt(result.viewTimes, 10)
		return &res, nil
	}

	return nil, nil
}
