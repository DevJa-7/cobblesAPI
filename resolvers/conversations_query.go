package resolvers

import (
	"context"
	"errors"
	"log"
	"strconv"

	"github.com/jackc/pgx"
)

type ConversationsInput struct {
	PostID *string

	PageToken *string
	Limit     *int32
}

type ConversationsResult struct {
	conversations []*ConversationResolver
	nextPageToken *string
}

func (r *Resolver) Conversations(ctx context.Context, req struct {
	Input *ConversationsInput
}) (*ConversationsResult, error) {
	if req.Input == nil {
		req.Input = &ConversationsInput{}
	}

	var limit int32
	if req.Input.Limit == nil || *req.Input.Limit == 0 || *req.Input.Limit > 1000 {
		// TODO: pagination, high limit for now
		limit = 1000
	} else {
		limit = *req.Input.Limit
	}

	userID, err := ctxUserID(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		select id, post_id, started_by_user_id, created_at, array_agg(user_id)
		from conversations
		join conversation_has_users on id = conversation_id
		where conversation_id in (
		select conversation_id
		from conversation_has_users
		where user_id=$1
			and (post_id=$2 or $2 is null)
		)
		group by id, started_by_user_id, created_at 
		order by id asc
		limit $3
	`

	var rows *pgx.Rows
	if req.Input.PostID != nil {
		postID, err := strconv.ParseInt(*req.Input.PostID, 10, 64)
		if err != nil {
			return nil, err
		}
		rows, err = r.server.ConnPool.Query(query, userID, postID, limit)
	} else {
		rows, err = r.server.ConnPool.Query(query, userID, nil, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []*ConversationResolver
	for rows.Next() {
		var c conversation
		err := rows.Scan(&c.id, &c.postID, &c.startedByUserID, &c.createdAt, &c.userIDs)
		if err != nil {
			return nil, err
		}

		log.Println("#####conversations ===", c)
		conversations = append(conversations, &ConversationResolver{
			resolver:     r,
			server:       r.server,
			conversation: &c,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &ConversationsResult{
		conversations: conversations,
	}, nil
}

func (r *ConversationsResult) Conversations() []*ConversationResolver {
	return r.conversations
}

func (r *ConversationsResult) NextPageToken() *string {
	// always nil, no pagination implemented
	return r.nextPageToken
}

func (r *Resolver) ConversationByID(ctx context.Context, args struct {
	ID string
}) (*ConversationResolver, error) {
	userID, err := ctxUserID(ctx)
	if err != nil {
		return nil, err
	}

	convoID, err := strconv.ParseInt(args.ID, 10, 64)
	if err != nil {
		return nil, err
	}

	conversation, exists, err := r.getConversation(getConversationInput{ID: convoID})
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("not found")
	}

	var found bool
	for _, cUser := range conversation.userIDs {
		if userID == cUser {
			found = true
			break
		}
	}

	if !found {
		return nil, errors.New("unauthorized")
	}

	return &ConversationResolver{
		resolver:     r,
		server:       r.server,
		conversation: conversation,
	}, nil
}
