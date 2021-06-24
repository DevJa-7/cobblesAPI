package resolvers

import (
	"fmt"
	"log"

	"github.com/jackc/pgx"
	sq "gopkg.in/Masterminds/squirrel.v1"
)

type getConversationInput struct {
	ID int64

	StartedByUserID int64
	PostID          int64
}

type conversationParticipant struct {
	conversationID int64
	userID         int64

	userName *string
}

func (r *Resolver) notifyConversationParticipants(conversationID int64, senderID int64) error {
	rows, err := r.server.ConnPool.Query(`
			select conversation_id, user_id, users.name
			from conversation_has_users
			left join users on users.id = user_id
			where conversation_id = $1
		`, conversationID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var conversationParticipants []*conversationParticipant
	for rows.Next() {
		var c conversationParticipant
		err := rows.Scan(&c.conversationID, &c.userID, &c.userName)
		if err != nil {
			return err
		}

		conversationParticipants = append(conversationParticipants, &c)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	senderName := "Somebody"
	for _, p := range conversationParticipants {
		if p.userID == senderID && p.userName != nil {
			senderName = *p.userName
			break
		}
	}

	for _, p := range conversationParticipants {
		if senderID == p.userID {
			continue
		}

		notifBody := fmt.Sprintf("%s replied to your post", senderName)
		if err := r.server.PublishNotificationToUser(p.userID, notifBody); err != nil {
			log.Println(err)
		}
	}

	return nil
}

func (r *Resolver) getConversation(in getConversationInput) (*conversation, bool, error) {
	stmt := newSelectBuilder("id", "post_id", "started_by_user_id", "created_at").From("conversations")

	if in.ID > 0 {
		stmt = stmt.Where(sq.Eq{"id": in.ID})
	} else {
		if in.StartedByUserID > 0 {
			stmt = stmt.Where(sq.Eq{"started_by_user_id": in.StartedByUserID})
		}

		if in.PostID > 0 {
			stmt = stmt.Where(sq.Eq{"post_id": in.PostID})
		}
	}

	selectSQL, selectArgs, err := stmt.ToSql()
	if err != nil {
		return nil, false, err
	}

	row := r.server.ConnPool.QueryRow(selectSQL, selectArgs...)
	convo, err := r.scanConversation(row)
	if err == pgx.ErrNoRows {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, err
	}

	sql, args, err := newSelectBuilder("user_id").
		From("conversation_has_users").
		Where(sq.Eq{"conversation_id": convo.id}).
		ToSql()
	if err != nil {
		return nil, false, err
	}

	rows, err := r.server.ConnPool.Query(sql, args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, false, err
		}

		convo.userIDs = append(convo.userIDs, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	return convo, true, nil
}
