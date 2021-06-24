package resolvers

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/jackc/pgx/pgtype"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/lambdacollective/cobbles-api/server"
)

func (r *Resolver) CurrentUser(ctx context.Context) (*UserResolver, error) {
	currentUserID, err := ctxUserID(ctx)
	if err != nil {
		return nil, err
	}

	user, err := r.server.UserByID(currentUserID)
	if err != nil {
		return nil, err
	}

	return &UserResolver{
		server: r.server,
		user:   user,
	}, nil
}

// OtherUser - Profile of another user
func (r *Resolver) OtherUser(ctx context.Context, args struct {
	ID int32
}) (*UserResolver, error) {

	user, err := r.server.UserByID(int64(args.ID))
	if err != nil {
		return nil, err
	}

	return &UserResolver{
		server: r.server,
		user:   user,
	}, nil
}

func (r *Resolver) UpdateUser(ctx context.Context, args struct {
	Input struct {
		Name     *string
		PhotoURL *string
		ZIPCode  *string
		Bio      *string
	}
}) (*UserResolver, error) {
	currentUserID, err := ctxUserID(ctx)
	if err != nil {
		return nil, err
	}

	inputName := args.Input.Name
	inputPhotoURL := args.Input.PhotoURL
	inputZIPCode := args.Input.ZIPCode
	inputBio := args.Input.Bio

	var u server.User
	var result struct {
		createdAt pgtype.Timestamptz
		updatedAt pgtype.Timestamptz
	}
	if err := r.server.ConnPool.QueryRow(`
		update users
		set name = coalesce($2, name),
			photo_url = coalesce($3, photo_url),
			zip_code = coalesce($4, zip_code),
			bio = coalesce($5, bio),
			updated_at = now()
		where id = $1
		returning
			id,
			name,
			phone_number,
			zip_code,
			bio,
			photo_url,
			created_at,
			updated_at
	`, currentUserID, inputName, inputPhotoURL, inputZIPCode, inputBio).Scan(
		&u.ID,
		&u.Name,
		&u.PhoneNumber,
		&u.ZIPCode,
		&u.Bio,
		&u.PhotoURL,
		&result.createdAt,
		&result.updatedAt,
	); err != nil {
		return nil, err
	}

	u.CreatedAt = result.createdAt.Time
	u.UpdatedAt = result.updatedAt.Time

	return &UserResolver{
		server: r.server,
		user:   &u,
	}, nil
}

func (r *Resolver) UpdateFCMToken(ctx context.Context, args struct {
	Input struct {
		UserID   string
		FCMToken string
	}
}) (*UserResolver, error) {

	inputUserID := args.Input.UserID
	inputFCMToken := args.Input.FCMToken

	var u server.User
	var result struct {
		createdAt pgtype.Timestamptz
		updatedAt pgtype.Timestamptz
	}
	if err := r.server.ConnPool.QueryRow(`
		update users
		set fcm_token = $2
		where id = $1
		returning
			id,
			name,
			phone_number,
			zip_code,
			bio,
			photo_url,
			fcm_token,
			created_at,
			updated_at
	`, inputUserID, inputFCMToken).Scan(
		&u.ID,
		&u.Name,
		&u.PhoneNumber,
		&u.ZIPCode,
		&u.Bio,
		&u.PhotoURL,
		&u.FCMToken,
		&result.createdAt,
		&result.updatedAt,
	); err != nil {
		return nil, err
	}

	u.CreatedAt = result.createdAt.Time
	u.UpdatedAt = result.updatedAt.Time

	return &UserResolver{
		server: r.server,
		user:   &u,
	}, nil
}

type UserResolver struct {
	server *server.Server

	user *server.User
}

type UsersResolver struct {
	server *server.Server
	users  []server.User
}

func (r *UserResolver) ID() graphql.ID {
	return graphql.ID(strconv.FormatInt(r.user.ID, 10))
}

func (r *UserResolver) Name() *string {
	return r.user.Name
}

func (r *UserResolver) PhoneNumber() *string {
	return r.user.PhoneNumber
}

func (r *UserResolver) PhotoURL() *string {
	if r.user.PhotoURL != nil {
		u, err := url.Parse(*r.user.PhotoURL)
		if err != nil {
			log.Print(err)
			return nil
		}

		imageURL := fmt.Sprintf("%s%s", r.server.ImgixUserMediaMediaEndpoint, u.Path)
		return &imageURL
	}

	return nil
}

func (r *UserResolver) ZIPCode() *string {
	return r.user.ZIPCode
}

func (r *UserResolver) Bio() *string {
	return r.user.Bio
}

func (r *UserResolver) Followers() *int32 {
	return r.user.Followers
}

func (r *UserResolver) Following() *int32 {
	return r.user.Following
}

func (r *UserResolver) PostCount() int32 {
	return r.user.PostCount
}

func (r *UserResolver) CreatedAt() Timestamp {
	return Timestamp{r.user.CreatedAt}
}

func (r *UserResolver) UpdatedAt() Timestamp {
	return Timestamp{r.user.UpdatedAt}
}

type OrderPostsBy string

const (
	OrderPostsChronologically OrderPostsBy = "CHRONOLOGICAL"
	OrderPostsBackwards       OrderPostsBy = "REVERSE_CHRONOLOGICAL"
)

type UserPostsInput struct {
	PageToken   *string
	Limit       *int32
	OtherUserID *int32
}

type UserPostsResult struct {
	posts         []*PostResolver
	nextPageToken *string
}

func (r *UserResolver) Posts(ctx context.Context, req struct {
	Input *UserPostsInput
}) (*UserPostsResult, error) {
	userID, err := ctxUserID(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: error codes?
	// if r.user.ID != userID {
	// 	return nil, errors.New("unauthorized")
	// }

	if req.Input.OtherUserID != nil {
		userID = int64(*req.Input.OtherUserID)
	}

	if req.Input == nil {
		req.Input = &UserPostsInput{}
	}

	if req.Input == nil {
		req.Input = &UserPostsInput{}
	}

	var limit int32
	if req.Input.Limit == nil || *req.Input.Limit == 0 || *req.Input.Limit > 100 {
		limit = 100
	} else {
		limit = *req.Input.Limit
	}

	postResolvers, nextPageToken, err := resolvePosts(ctx, r.server, resolvePostsInput{
		ByUserID: userID,

		PageToken: req.Input.PageToken,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}

	return &UserPostsResult{
		posts:         postResolvers,
		nextPageToken: nextPageToken,
	}, nil
}

func (r *UserPostsResult) Posts() []*PostResolver {
	return r.posts
}

func (r *UserPostsResult) NextPageToken() *string {
	return r.nextPageToken
}
