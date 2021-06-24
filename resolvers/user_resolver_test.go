package resolvers

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserPosts(t *testing.T) {
	harness := NewTestHarness(t)
	harness.ResetDB()

	for i := 0; i < 6; i++ {
		iStr := strconv.Itoa(i)
		vars := map[string]interface{}{
			"input": map[string]interface{}{
				"title":       "title " + iStr,
				"description": "desc " + iStr,
				"kind":        "TEXT",
				"poster":      "nan",
			},
		}

		// good
		var userID int64 = 1
		if i == 3 {
			// no good
			userID = 2
		}

		harness.MustCreateUser(userID)

		harness.MustExec(ExecInput{
			Variables: vars,
			UserID:    userID,
			Query: `
				mutation CreatePost($input: CreatePostInput!) {
					createPost(input: $input) {
						id
						title
						description
						kind
						poster
					}
				}
			`}, nil)
	}

	t.Run("basic", func(t *testing.T) {
		harness := NewTestHarness(t)

		for _, args := range []string{"", "input: {}"} {
			t.Run(fmt.Sprintf("args: %s", args), func(t *testing.T) {
				harness.GQLAssert("should return posts", GQLAssertInput{
					ExecInput: ExecInput{
						UserID: 1,
						Query: fmt.Sprintf(`
						{
							currentUser() {
								posts(%s) {
									posts {
										title
									}
								}
							}
						}`, args),
					},
					ExpectedResult: map[string]interface{}{
						"currentUser": map[string]interface{}{
							"posts": map[string]interface{}{
								"posts": []interface{}{
									map[string]interface{}{
										"title": "title 5",
									},
									map[string]interface{}{
										"title": "title 4",
									},
									// 3 is user 2's, not 1's
									map[string]interface{}{
										"title": "title 2",
									},
									map[string]interface{}{
										"title": "title 1",
									},
									map[string]interface{}{
										"title": "title 0",
									},
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
			UserID: 1,
			Query: `
			{
				currentUser() {
					posts(input: {limit: 2}) {
						posts {
							title
						}
						nextPageToken
					}
				}
			}
			`,
		}, &res1)

		extractResp := func(r map[string]interface{}) map[string]interface{} {
			return r["currentUser"].(map[string]interface{})["posts"].(map[string]interface{})
		}

		pageToken1 := extractResp(res1)["nextPageToken"].(string)
		require.NotEmpty(t, pageToken1)

		posts1 := extractResp(res1)["posts"].([]interface{})
		require.Equal(t, "title 5", posts1[0].(map[string]interface{})["title"])
		require.Equal(t, "title 4", posts1[1].(map[string]interface{})["title"])

		var res2 map[string]interface{}
		harness.MustExec(ExecInput{
			UserID: 1,
			Query: fmt.Sprintf(`
			{
				currentUser() {
					posts(input: {pageToken: "%s", limit: 2}) {
						posts {
							title
						}
						nextPageToken
					}
				}
			}
			`, pageToken1),
		}, &res2)

		pageToken2 := extractResp(res2)["nextPageToken"].(string)
		require.NotEmpty(t, pageToken2)

		posts2 := extractResp(res2)["posts"].([]interface{})
		require.Equal(t, "title 2", posts2[0].(map[string]interface{})["title"])
		require.Equal(t, "title 1", posts2[1].(map[string]interface{})["title"])

		var res3 map[string]interface{}
		harness.Exec(ExecInput{
			UserID: 1,
			Query: fmt.Sprintf(`
			{
				currentUser() {
					posts(input: {limit: 2, pageToken: "%s"}) {
						posts {
							title
						}
						nextPageToken
					}
				}
			}`, pageToken2),
		}, &res3)
		pageToken3 := extractResp(res3)["nextPageToken"]
		require.Nil(t, pageToken3)
	})
}
