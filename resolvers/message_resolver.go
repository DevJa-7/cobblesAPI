package resolvers

import (
	"strconv"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/lambdacollective/cobbles-api/server"
)

type MessageResolver struct {
	resolver *Resolver
	server   *server.Server

	message      *message
	conversation *conversation
	from         *server.User
}

func (m *MessageResolver) ID() graphql.ID {
	return graphql.ID(strconv.FormatInt(m.message.id, 10))
}

func (m *MessageResolver) From() *UserResolver {
	return &UserResolver{m.server, m.from}
}

func (m *MessageResolver) Conversation() *ConversationResolver {
	return &ConversationResolver{
		resolver:     m.resolver,
		server:       m.server,
		conversation: m.conversation,
	}
}

func (m *MessageResolver) Body() string {
	return m.message.body
}

func (m *MessageResolver) Timestamp() Timestamp {
	return Timestamp{m.message.createdAt}
}

type message struct {
	id             int64
	fromUserID     int64
	conversationID int64
	body           string
	createdAt      time.Time
}
