package resolvers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

type GetOrCreateConversationInput struct {
	PostID string `validate:"required"`
}

type GetOrCreateConversationResult struct {
	conversation *ConversationResolver
	created      bool
}

func (r *Resolver) GetOrCreateConversation(ctx context.Context, req struct {
	Input *GetOrCreateConversationInput
}) (*GetOrCreateConversationResult, error) {
	if req.Input == nil {
		req.Input = &GetOrCreateConversationInput{}
	}

	if err := validate.Struct(req); err != nil {
		return nil, err
	}

	userID, err := ctxUserID(ctx)
	if err != nil {
		return nil, err
	}

	postID, err := strconv.ParseInt(req.Input.PostID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid post ID: %s", req.Input.PostID)
	}

	post, exists, err := r.getPost(postID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get post")
	}

	if !exists {
		return nil, errors.New("post not found")
	}

	if post.UserID == userID {
		return nil, errors.New("can't message your own post")
	}

	convo, exists, err := r.getConversation(getConversationInput{
		StartedByUserID: userID,
		PostID:          postID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "get conversation")
	}

	if exists {
		return &GetOrCreateConversationResult{
			conversation: &ConversationResolver{
				server:       r.server,
				conversation: convo,
				post:         post,
			},
			created: false,
		}, nil
	}

	convo, err = r.insertConversation(post.ID, userID, post.UserID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert convo")
	}

	return &GetOrCreateConversationResult{
		conversation: &ConversationResolver{
			server:       r.server,
			conversation: convo,
			post:         post,
		},
		created: true,
	}, nil
}

func (r *Resolver) insertConversation(postID, userID, otherUserID int64) (*conversation, error) {
	tx, err := r.server.ConnPool.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Inserts have side-effects like incrementing sequences so I just always select first
	// (even though there's a potential but harmless race)
	sql, args, err := newInsertBuilder("conversations").
		Columns("post_id", "started_by_user_id").
		Values(postID, userID).
		Suffix("RETURNING id, post_id, started_by_user_id, created_at").
		ToSql()
	if err != nil {
		return nil, err
	}

	row := r.server.ConnPool.QueryRow(sql, args...)
	convo, err := r.scanConversation(row)
	if err != nil {
		return nil, errors.Wrap(err, "failed to scan convo")
	}

	sql, args, err = newInsertBuilder("conversation_has_users").
		Columns("conversation_id", "user_id").
		Values(convo.id, userID).
		Values(convo.id, otherUserID).
		ToSql()
	if err != nil {
		return nil, err
	}

	_, err = r.server.ConnPool.Exec(sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert conversation:user mapping")
	}

	convo.userIDs = []int64{userID, otherUserID}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return convo, nil
}

func (r *GetOrCreateConversationResult) Conversation() *ConversationResolver {
	return r.conversation
}

func (r *GetOrCreateConversationResult) Created() bool {
	return r.created
}

func (r *Resolver) scanConversation(row scannable) (*conversation, error) {
	var c conversation
	err := row.Scan(&c.id, &c.postID, &c.startedByUserID, &c.createdAt)
	return &c, err
}
