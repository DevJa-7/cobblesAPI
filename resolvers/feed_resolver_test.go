package resolvers

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFeedResolver(t *testing.T) {
	harness := NewTestHarness(t)
	harness.ResetDB()

	for i := 0; i < 6; i++ {
		iStr := strconv.Itoa(i)

		inputVars := map[string]interface{}{
			"title":       "title " + iStr,
			"description": "desc " + iStr,
			"kind":        "TEXT",
			"poster":      "default",
		}

		// some posts get tags, i may go up to higher later
		// index is depended on by tests
		switch i {
		case 1:
			inputVars["tags"] = []interface{}{"free"}
		case 2:
			inputVars["tags"] = []interface{}{"for sale"}
		case 3:
			inputVars["tags"] = []interface{}{"for sale", "crime"}
		case 4:
			inputVars["kind"] = "IMAGE"
			inputVars["mediaURL"] = "http://llc-cobbles-dev-user-media.s3-external-1.amazonaws.com/images/some_image"
		case 5:
			inputVars["kind"] = "IMAGE"
			inputVars["mediaURL"] = "http://llc-cobbles-dev-user-media.s3-external-1.amazonaws.com/images/some_image"
			inputVars["mediaMetadata"] = map[string]interface{}{
				"width": 111,
				// "height"
			}
		}

		harness.MustExec(ExecInput{
			Variables: map[string]interface{}{
				"input": inputVars,
			},
			Query: `
				mutation CreatePost($input: CreatePostInput!) {
					createPost(input: $input) {
						id
					}
				}
			`}, nil)
	}

	t.Run("works", func(t *testing.T) {
		harness := NewTestHarness(t)

		for _, args := range []string{"", "input: {}"} {
			t.Run(fmt.Sprintf("args: %s", args), func(t *testing.T) {
				harness.GQLAssert("should return posts", GQLAssertInput{
					ExecInput: ExecInput{
						Query: fmt.Sprintf(`
						{
							feed(%s) {
								posts {
									title
									description
									kind
									media {
										url
										width
										height
									}
								}
							}
						}`, args),
					},
					ExpectedResult: map[string]interface{}{
						"feed": map[string]interface{}{
							"posts": []map[string]interface{}{
								{
									"title":       "title 5",
									"description": "desc 5",
									"kind":        "IMAGE",
									"media": map[string]interface{}{
										"url":    "https://llc-cobbles-dev-user-images.imgix.net/images/some_image",
										"width":  111,
										"height": nil,
									},
								},
								{
									"title":       "title 4",
									"description": "desc 4",
									"kind":        "IMAGE",
									"media": map[string]interface{}{
										"url":    "https://llc-cobbles-dev-user-images.imgix.net/images/some_image",
										"width":  nil,
										"height": nil,
									},
								},
								{
									"title":       "title 3",
									"description": "desc 3",
									"kind":        "TEXT",
									"media":       nil,
								},
								{
									"title":       "title 2",
									"description": "desc 2",
									"kind":        "TEXT",
									"media":       nil,
								},
								{
									"title":       "title 1",
									"description": "desc 1",
									"kind":        "TEXT",
									"media":       nil,
								},
								{
									"title":       "title 0",
									"description": "desc 0",
									"kind":        "TEXT",
									"media":       nil,
								},
							},
						},
					},
				})
			})
		}
	})

	t.Run("pagination", func(t *testing.T) {
		harness := NewTestHarness(t)

		var res1 map[string]interface{}
		harness.MustExec(ExecInput{
			Query: `
			{
				feed(input: {limit: 2}) {
					posts {
						title
					}
					nextPageToken
				}
			}
			`,
		}, &res1)

		pageToken1 := res1["feed"].(map[string]interface{})["nextPageToken"].(string)
		require.NotEmpty(t, pageToken1)

		posts1 := res1["feed"].(map[string]interface{})["posts"].([]interface{})
		require.Equal(t, "title 5", posts1[0].(map[string]interface{})["title"])
		require.Equal(t, "title 4", posts1[1].(map[string]interface{})["title"])

		var res2 map[string]interface{}
		harness.MustExec(ExecInput{
			Query: fmt.Sprintf(`
			{
				feed(input: {pageToken: "%s", limit: 2}) {
					posts {
						title
					}
					nextPageToken
				}
			}
			`, pageToken1),
		}, &res2)

		pageToken2 := res2["feed"].(map[string]interface{})["nextPageToken"].(string)
		require.NotEmpty(t, pageToken2)

		posts2 := res2["feed"].(map[string]interface{})["posts"].([]interface{})
		require.Equal(t, "title 3", posts2[0].(map[string]interface{})["title"])
		require.Equal(t, "title 2", posts2[1].(map[string]interface{})["title"])

		var res3 map[string]interface{}
		harness.Exec(ExecInput{
			Query: fmt.Sprintf(`
			{
				feed(input: {limit: 2, pageToken: "%s"}) {
					nextPageToken
				}
			}`, pageToken2),
		}, &res3)
		pageToken3 := res3["feed"].(map[string]interface{})["nextPageToken"]
		require.Nil(t, pageToken3)
	})

	t.Run("tags", func(t *testing.T) {
		harness := NewTestHarness(t)
		harness.GQLAssert("should select posts with ANY matching tag", GQLAssertInput{
			ExecInput: ExecInput{
				Query: `
				{
					feed(input: {tags: ["for sale"]}) {
						posts {
							title
							tags
						}
					}
				}`,
			},
			ExpectedResult: map[string]interface{}{
				"feed": map[string]interface{}{
					"posts": []map[string]interface{}{
						// some overlap
						{
							"title": "title 3",
							"tags":  []interface{}{"for sale", "crime"},
						},
						// exact match
						{
							"title": "title 2",
							"tags":  []interface{}{"for sale"},
						},
						// post 2 has tags but no overlap
						// rest of posts have no tags
					},
				},
			},
		})
	})
}
