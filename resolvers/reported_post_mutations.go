package resolvers

import (
	"context"
)

type ReportedPostInput struct {
	PostID              int32
	PersonReportingID   int32
	ReportingReasonText string
}

func (r *Resolver) CreateReportedPost(ctx context.Context, args struct {
	Input ReportedPostInput
}) (*ReportedPostResolver, error) {
	reportedPost, err := r.server.CreateReportedPost(
		args.Input.PostID,
		args.Input.PersonReportingID,
		args.Input.ReportingReasonText,
	)

	if err != nil {
		return nil, err
	}

	return &ReportedPostResolver{server: r.server, reportedPost: reportedPost}, nil
}
