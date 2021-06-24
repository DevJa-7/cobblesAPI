package resolvers

import (
	"log"
	"strconv"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/lambdacollective/cobbles-api/server"
)

type PostComment struct {
	ParentCommentID int32
	Comment         string
	ID              int32
}

// CreatePostCommentInput ...
type CreatePostCommentInput struct {
	ParentCommentID int32
	Comment         string
	PostID          int32
}

// RemovePostCommentInput ...
type RemovePostCommentInput struct {
	CommentID int32
	PostID    int32
}

type PostCommentsResult struct {
	comments      []*PostCommentResolver
	nextPageToken *string
}

// original code. reason: PostCommentResolver redeclared.
// type PostCommentResolver struct {
// 	server      *server.Server
// 	postComment PostComment
// }

// func (r *PostCommentResolver) ID() graphql.ID {
// 	return graphql.ID(r.postComment.ID)
// }

func (r *PostCommentResolver) ParentCommentID() graphql.ID {
	return graphql.ID(r.postComment.ParentCommentID)
}

// UserPostCommentsInput : post comments input
type UserPostCommentsInput struct {
	PageToken *string
	Limit     *int32
	// OtherUserID *int32
	PostID int32
}

// PostCommentResolver : post comment resolver
type PostCommentResolver struct {
	server *server.Server

	postComment *server.PostComment
}

// ID : ID Resolver of the post comment
func (r *PostCommentResolver) ID() graphql.ID {
	return graphql.ID(strconv.FormatInt(r.postComment.ID, 10))
}

// Author : Author Resolver of the post comment
func (r *PostCommentResolver) Author() (*UserResolver, error) {
	user, err := r.server.UserByID(int64(r.postComment.UserID))
	if err != nil {
		log.Println(err)
		return nil, nil
	}
	return &UserResolver{
		server: r.server,
		user:   user,
	}, nil
}

// Comment : Comment Resolver of the post comment
func (r *PostCommentResolver) Comment() string {
	return r.postComment.Comment
}

func (r *PostCommentsResult) Comments() []*PostCommentResolver {
	return r.comments
}

// func (r *PostCommentResolver) ParentCommentID() graphql.ID {
// 	return graphql.ID(int64(r.postComment.ParentCommentID))
// }

// func (r *PostCommentResolver) PostID() graphql.ID {
// 	return graphql.ID(int64(r.postComment.PostID))
// }

// CreatedAt : Create At Resolver of the post comments
func (r *PostCommentResolver) CreatedAt() Timestamp {
	return Timestamp{r.postComment.CreatedAt}
}

// UpdatedAt : Update At Resolver of the post comments
func (r *PostCommentResolver) UpdatedAt() Timestamp {
	return Timestamp{r.postComment.UpdatedAt}
}

// PostComments : postComments Resolver
func (r *Resolver) PostComments(req struct {
	Input *UserPostCommentsInput
}) (*PostCommentsResult, error) {

	// if req.Input.OtherUserID != nil {
	// 	userID = int64(*req.Input.OtherUserID)
	// }

	if req.Input == nil {
		req.Input = &UserPostCommentsInput{}
	}

	var limit int32
	if req.Input.Limit == nil || *req.Input.Limit == 0 || *req.Input.Limit > 100 {
		limit = 100
	} else {
		limit = *req.Input.Limit
	}

	postCommentResolvers, nextPageToken, err := resolvePostComments(r.server, resolvePostCommentsInput{
		// ByUserID: userID,
		PostID:    req.Input.PostID,
		PageToken: req.Input.PageToken,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}

	return &PostCommentsResult{
		comments:      postCommentResolvers,
		nextPageToken: nextPageToken,
	}, nil
}

// PostComments : Post comment Resolver
func (r *PostCommentsResult) PostComments() []*PostCommentResolver {
	return r.comments
}

// NextPageToken : Next page token Resolver of the post comments
func (r *PostCommentsResult) NextPageToken() *string {
	return r.nextPageToken
}
