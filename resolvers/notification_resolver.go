package resolvers

import (
	"strconv"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/lambdacollective/cobbles-api/server"
)

type notification struct {
	ID        int64
	Unread    bool
	Timestamp time.Time
	CreatedAt time.Time
}

type NotificationResolver struct {
	server       *server.Server
	notification *notification
}

func (r *NotificationResolver) ID() graphql.ID {
	return graphql.ID(strconv.FormatInt(r.notification.ID, 10))
}

func (r *NotificationResolver) Content() string {
	return ""
}

func (r *NotificationResolver) Unread() bool {
	return r.notification.Unread
}

func (r *NotificationResolver) Timestamp() Timestamp {
	return Timestamp{}
}
