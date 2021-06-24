package resolvers

import (
	"fmt"
	"log"
	"strconv"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/lambdacollective/cobbles-api/server"
)

const (
	PostKindText  server.PostKind = "TEXT"
	PostKindImage server.PostKind = "IMAGE"
	PostKindVideo server.PostKind = "VIDEO"
)

type PostMedia struct {
	URL    string `json:"url,omitempty"`
	Width  *int32 `json:"width,omitempty"`
	Height *int32 `json:"height,omitempty"`
}

type PostResolver struct {
	server *server.Server

	post         *server.Post
	neighborhood *server.Neighborhood
}

func (r *PostResolver) ID() graphql.ID {
	return graphql.ID(strconv.FormatInt(r.post.ID, 10))
}

func (r *PostResolver) Title() string {
	return r.post.Title
}

func (r *PostResolver) Description() *string {
	return r.post.Description
}

func (r *PostResolver) Kind() *server.PostKind {
	return &r.post.Kind
}

func (r *PostResolver) Poster() *string {
	return r.post.Poster
}

func (r *PostResolver) Author() (*UserResolver, error) {
	user, err := r.server.UserByID(r.post.UserID)
	if err != nil {
		log.Println(err)
		return nil, nil
	}

	return &UserResolver{
		server: r.server,
		user:   user,
	}, nil
}

func (r *PostResolver) Neighborhood() (*NeighborhoodResolver, error) {
	if r.neighborhood != nil {
		return &NeighborhoodResolver{neighborhood: r.neighborhood}, nil
	}

	neighborhood, err := r.server.NeighborhoodByID(r.post.NeighborhoodID)
	if err != nil {
		log.Println(err)
		return nil, nil
	}

	return &NeighborhoodResolver{
		server:       r.server,
		neighborhood: neighborhood,
	}, nil
}

func (r *PostResolver) Preview() *PostMediaResolver {
	if r.post.Preview != nil {
		return &PostMediaResolver{
			server: r.server,

			post:      r.post,
			postMedia: r.post.Preview,
			isImage:   true,
		}
	}

	return nil
}

func (r *PostResolver) Media() *PostMediaResolver {
	if r.post.Media != nil {
		return &PostMediaResolver{
			server: r.server,

			post:      r.post,
			postMedia: r.post.Media,
		}
	}

	return nil
}

func (r *PostResolver) Processing() bool {
	return r.post.Processing
}

func (r *PostResolver) ShareLinkURL() *string {
	u := fmt.Sprintf("http://example.com/posts/%d", r.post.ID)
	return &u
}

func (r *PostResolver) Tags() *[]string {
	return &r.post.Tags
}

func (r *PostResolver) CreatedAt() Timestamp {
	return Timestamp{r.post.CreatedAt}
}

func (r *PostResolver) UpdatedAt() Timestamp {
	return Timestamp{r.post.UpdatedAt}
}

func (r *PostResolver) Likes() *int32 {
	return r.post.Likes
}

func (r *PostResolver) CommentCount() *int32 {
	return r.post.CommentCount
}

func (r *PostResolver) ViewTimes() *string {
	viewTimes := r.post.ViewTimes
	res := strconv.FormatInt(*viewTimes, 10)
	return &res
}
