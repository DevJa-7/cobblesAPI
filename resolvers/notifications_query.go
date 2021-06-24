package resolvers

import (
	"context"
	"errors"
)

type NotificationsInput struct {
	KeepSchemaHappy *bool
}

func (r *Resolver) Notifications(ctx context.Context, args struct{ Input *NotificationsInput }) (*NotificationsResultResolver, error) {
	return nil, errors.New("not implemented")

	currentUserID, err := ctxUserID(ctx)
	if err != nil {
		return nil, err
	}

	r.server.ConnPool.Query(`
		select 
			id
			data
			created_at
		from notifications
		where user_id = $1
		order by created_at desc
	`, currentUserID)

	return &NotificationsResultResolver{}, nil
}

type NotificationsResultResolver struct{}

func (r *NotificationsResultResolver) Notifications() []*NotificationResolver {
	return []*NotificationResolver{}
}
