package resolvers

import (
	"context"
	"log"
	"strconv"

	"github.com/pkg/errors"
)

type SendMessageInput struct {
	ConversationID    string `validate:"required"`
	DestinationUserID string `validate:"required"`
	Body              string `validate:"required"`
}

type SendMessageResult struct {
	message *MessageResolver
}

func (r *Resolver) SendMessage(ctx context.Context, in struct {
	Input *SendMessageInput
}) (*SendMessageResult, error) {
	req := in.Input

	if err := validate.Struct(req); err != nil {
		return nil, err
	}

	userID, err := ctxUserID(ctx)
	if err != nil {
		return nil, err
	}

	convoID, err := strconv.ParseInt(req.ConversationID, 10, 64)
	if err != nil {
		return nil, err
	}

	convo, err := r.validateConvoOwnership(userID, convoID)
	if err != nil {
		return nil, err
	}

	sql, args, err := newInsertBuilder("messages").
		Columns("from_user_id", "conversation_id", "body").
		Values(userID, convoID, req.Body).
		Suffix("RETURNING id, from_user_id, conversation_id, body, created_at").
		ToSql()
	if err != nil {
		return nil, err
	}

	row := r.server.ConnPool.QueryRow(sql, args...)
	msg, err := r.scanMessage(row)
	if err != nil {
		return nil, err
	}

	if err := r.notifyConversationParticipants(convoID, userID); err != nil {
		log.Println(err)
	}

	user, err := r.server.UserByID(userID)
	if err != nil {
		return nil, err
	}

	destUserID, err := strconv.ParseInt(req.DestinationUserID, 10, 64)
	if err != nil {
		return nil, err
	}

	var fcmToken string
	// get the destination user's fcm token
	err = r.server.ConnPool.QueryRow(`
		select fcm_token from users where id = $1
	`, destUserID).Scan(
		&fcmToken,
	)
	if err != nil {
		return nil, err
	}

	// send notification
	msgType := "messaging"
	if msg.id > 0 {
		err = r.server.SendNotification(fcmToken, msgType, msg.id, req.Body, *user.Name, convoID, userID)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	return &SendMessageResult{
		message: &MessageResolver{
			resolver:     r,
			server:       r.server,
			message:      msg,
			conversation: convo,
			from:         user,
		},
	}, nil
}

func (r *Resolver) scanMessage(row scannable) (*message, error) {
	var m message
	err := row.Scan(&m.id, &m.fromUserID, &m.conversationID, &m.body, &m.createdAt)
	return &m, err
}

func (r *Resolver) validateConvoOwnership(userID int64, convoID int64) (*conversation, error) {
	convo, exists, err := r.getConversation(getConversationInput{ID: convoID})
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("convo: not found")
	}

	var found bool
	for _, convoUserID := range convo.userIDs {
		if convoUserID == userID {
			found = true
		}
	}

	if !found {
		return nil, errors.New("unauthorized")
	}

	return convo, nil
}

func (s *SendMessageResult) Message() *MessageResolver {
	return s.message
}
