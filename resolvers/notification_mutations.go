package resolvers

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
)

type MarkNotificationReadInput struct {
	NotificationID graphql.ID
}

func (r *Resolver) MarkNotificationRead(ctx context.Context, args struct {
	Input MarkNotificationReadInput
}) (*bool, error) {
	currentUserID, err := ctxUserID(ctx)
	if err != nil {
		return nil, err
	}

	r.server.ConnPool.Query(`
		update notifications
		set read = true
		where user_id = $1
			and id = $2
	`, currentUserID, args.Input.NotificationID)

	return nil, nil
}
