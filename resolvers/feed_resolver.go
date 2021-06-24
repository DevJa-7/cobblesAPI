package resolvers

import (
	"context"
)

type FeedInput struct {
	PageToken *string
	Tags      *[]string
	Limit     *int32
}

type FeedResult struct {
	posts         []*PostResolver
	nextPageToken *string
}

func (r *Resolver) Feed(ctx context.Context, req struct {
	Input *FeedInput
}) (*FeedResult, error) {
	if req.Input == nil {
		req.Input = &FeedInput{}
	}

	var limit int32
	if req.Input.Limit == nil || *req.Input.Limit == 0 || *req.Input.Limit > 100 {
		limit = 100
	} else {
		limit = *req.Input.Limit
	}

	postResolvers, nextPageToken, err := resolvePosts(ctx, r.server, resolvePostsInput{
		// like a twitter feed, most recent at top
		Tags:      req.Input.Tags,
		PageToken: req.Input.PageToken,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}

	return &FeedResult{
		posts:         postResolvers,
		nextPageToken: nextPageToken,
	}, nil
}

func (f *FeedResult) Posts() []*PostResolver {
	return f.posts
}

func (f *FeedResult) NextPageToken() *string {
	return f.nextPageToken
}
