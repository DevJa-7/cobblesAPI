package resolvers

import (
	"database/sql"
	"strconv"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/lambdacollective/cobbles-api/server"
)

type ReportedPost struct {
	ID                  int64
	PostID              int64
	PersonReportingID   int64
	ReportingReasonText string
	ActionTaken         sql.NullInt64
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type ReportedPostResolver struct {
	server       *server.Server
	reportedPost *server.ReportedPost
}

func (r *ReportedPostResolver) ID() graphql.ID {
	return graphql.ID(strconv.FormatUint(uint64(r.reportedPost.ID), 10))
}

func (r *ReportedPostResolver) PostID() int32 {
	return int32(r.reportedPost.PostID)
}

func (r *ReportedPostResolver) PersonReportingID() int32 {
	return int32(r.reportedPost.PersonReportingID)
}

func (r *ReportedPostResolver) ReportingReasonText() string {
	return r.reportedPost.ReportingReasonText
}

func (r *ReportedPostResolver) ActionTaken() int32 {
	// r.reportedPost.ActionTaken.
	// actionStr := strconv.Itoa(r.reportedPost.ActionTaken)
	return int32(r.reportedPost.ActionTaken.Int64)
}

func (r *ReportedPostResolver) CreatedAt() Timestamp {
	return Timestamp{r.reportedPost.CreatedAt}
}

func (r *ReportedPostResolver) UpdatedAt() Timestamp {
	return Timestamp{r.reportedPost.UpdatedAt}
}
