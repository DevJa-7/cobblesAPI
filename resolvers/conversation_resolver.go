package resolvers

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/lambdacollective/cobbles-api/server"
	sq "gopkg.in/Masterminds/squirrel.v1"
)

type ConversationResolver struct {
	resolver     *Resolver
	server       *server.Server
	conversation *conversation
	post         *server.Post
}

type conversation struct {
	id              int64
	postID          int64
	startedByUserID int64
	userIDs         []int64
	createdAt       time.Time
}

func (c *ConversationResolver) ID() graphql.ID {
	return graphql.ID(strconv.FormatInt(c.conversation.id, 10))
}

func (c *ConversationResolver) Post() (*PostResolver, error) {
	if c.post != nil {
		// ugh we should use dataloader or something, annoying to propagate everything
		// im just gonna let neighborhoods n+1 for now _if_ theyre selected
		return &PostResolver{c.server, c.post, nil}, nil
	}

	post, exists, err := c.resolver.getPost(c.conversation.postID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("post not found")
	}

	return &PostResolver{c.server, post, nil}, nil
}

func (c *ConversationResolver) StartedBy() (*UserResolver, error) {
	user, err := c.server.UserByID(c.conversation.startedByUserID)
	if err != nil {
		return nil, err
	}

	return &UserResolver{c.server, user}, nil
}

func (c *ConversationResolver) Participants() ([]*UserResolver, error) {
	resolvers := make([]*UserResolver, 0, 2)

	if len(c.conversation.userIDs) > 0 {
		for _, userID := range c.conversation.userIDs {
			user, err := c.server.UserByID(userID)
			if err != nil {
				return nil, err
			}

			resolvers = append(resolvers, &UserResolver{c.server, user})
		}
	} else {
		sql, args, err := newSelectBuilder("user_id").
			From("conversation_has_users").
			Where(sq.Eq{"conversation_id": c.conversation.id}).
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := c.server.ConnPool.Query(sql, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var userID int64
			if err := rows.Scan(&userID); err != nil {
				return nil, err
			}

			// This should be a JOIN with a generic scanUser
			user, err := c.server.UserByID(userID)
			if err != nil {
				return nil, err
			}

			resolvers = append(resolvers, &UserResolver{c.server, user})
		}

		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	// make response deterministic for tests
	// easier for now to keep everything ordered
	sort.Slice(resolvers, func(i, j int) bool {
		return resolvers[i].user.ID < resolvers[j].user.ID
	})

	return resolvers, nil
}

func (c *ConversationResolver) CreatedAt() Timestamp {
	return Timestamp{c.conversation.createdAt}
}

type ConversationMessagesInput struct {
	PageToken *string
	Limit     *int32
}

type ConversationMessagesResult struct {
	messages      []*MessageResolver
	nextPageToken *string
}

func (c *ConversationResolver) Messages(ctx context.Context, req struct {
	Input *ConversationMessagesInput
}) (*ConversationMessagesResult, error) {
	userID, err := ctxUserID(ctx)
	if err != nil {
		return nil, err
	}

	var found bool
	for _, convoUser := range c.conversation.userIDs {
		if convoUser == userID {
			found = true
			break
		}
	}

	if !found {
		return nil, errors.New("unauthorized")
	}

	if req.Input == nil {
		req.Input = &ConversationMessagesInput{}
	}

	var limit int32
	if req.Input.Limit == nil || *req.Input.Limit == 0 || *req.Input.Limit > 100 {
		limit = 100
	} else {
		limit = *req.Input.Limit
	}

	stmt := newSelectBuilder("id", "from_user_id", "conversation_id", "body", "created_at").
		From("messages").
		Where(sq.Eq{"conversation_id": c.conversation.id}).
		OrderBy("created_at desc").
		Limit(uint64(limit + 1))

	afterID, err := DecodeAfterIDCursor(req.Input.PageToken)
	if err != nil {
		return nil, err
	}

	if afterID > 0 {
		// less than because paginating in reverse chronological
		stmt = stmt.Where(sq.Lt{"id": afterID})
	}

	sql, args, err := stmt.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := c.server.ConnPool.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgResolvers []*MessageResolver
	var lastID int64
	var i int32
	for rows.Next() {
		if i == limit {
			lastID = msgResolvers[len(msgResolvers)-1].message.id
			break
		}

		msg, err := c.resolver.scanMessage(rows)
		if err != nil {
			return nil, err
		}

		user, err := c.server.UserByID(msg.fromUserID)
		if err != nil {
			return nil, err
		}

		msgResolvers = append(msgResolvers, &MessageResolver{
			server:   c.server,
			resolver: c.resolver,
			message:  msg,
			from:     user,
		})

		i++
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &ConversationMessagesResult{
		messages:      msgResolvers,
		nextPageToken: EncodeAfterIDCursor(lastID),
	}, nil
}

func (r *ConversationMessagesResult) Messages() []*MessageResolver {
	return r.messages
}

func (r *ConversationMessagesResult) NextPageToken() *string {
	return r.nextPageToken
}
