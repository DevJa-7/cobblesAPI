package resolvers

import (
	"fmt"
	"testing"

	"github.com/graph-gophers/graphql-go/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChat(t *testing.T) {
	harness := NewTestHarness(t)
	harness.ResetDB()

	var (
		user1ID string = "1"
		user2ID string = "2"
		user3ID string = "3"

		user1IDi64 int64 = 1
		user2IDi64 int64 = 2
		user3IDi64 int64 = 3

		// post IDs
		// user1Post3 does not have a conversation
		user1Post1ID, user1Post2ID, user1Post3ID string
		user2Post1ID                             string
		user3Post1ID                             string

		// conversation IDs
		user1Post1ConversationID, user1Post2ConversationID  string
		user2Post1ConversationID, user2Post1Conversation2ID string
		// the only conversation user ID 1 doesnt belong to
		user3Post1ConversationID string

		templateData map[string]interface{}
	)

	harness.MustCreateUser(user1IDi64)
	harness.MustCreateUser(user2IDi64)
	harness.MustCreateUser(user3IDi64)

	{
		makeConversation := func(userID int64, postID string) string {
			rawRes := map[string]interface{}{}
			harness.MustExec(ExecInput{
				UserID: userID,
				Query: fmt.Sprintf(`
				mutation {
					getOrCreateConversation(input: {
						postID: "%s"
					}) {
						conversation {
							id
						}
						created
					}
				 }`, postID),
			}, &rawRes)

			res := rawRes["getOrCreateConversation"].(map[string]interface{})

			id := res["conversation"].(map[string]interface{})["id"].(string)
			require.NotEmpty(t, id)

			created := res["created"].(bool)
			require.True(t, created)

			return id
		}

		makePost := func(userID int64) string {
			var res map[string]interface{}
			harness.MustExec(ExecInput{
				UserID: userID,
				Variables: map[string]interface{}{
					"input": map[string]interface{}{
						"title":       "title",
						"description": "desc",
						"kind":        "TEXT",
						"poster":      "default",
					},
				},
				Query: `
				mutation CreatePost($input: CreatePostInput!) {
					createPost(input: $input) {
						id
					}
				}
				`,
			}, &res)

			postID := res["createPost"].(map[string]interface{})["id"].(string)
			require.NotEmpty(t, postID)
			return postID
		}

		// posts
		user1Post1ID = makePost(user1IDi64)
		user1Post2ID = makePost(user1IDi64)
		user1Post3ID = makePost(user1IDi64)

		user2Post1ID = makePost(user2IDi64)

		user3Post1ID = makePost(user3IDi64)

		// convos
		user1Post1ConversationID = makeConversation(user2IDi64, user1Post1ID)
		user1Post2ConversationID = makeConversation(user2IDi64, user1Post2ID)

		user2Post1ConversationID = makeConversation(user1IDi64, user2Post1ID)
		user2Post1Conversation2ID = makeConversation(user3IDi64, user2Post1ID)

		user3Post1ConversationID = makeConversation(user2IDi64, user3Post1ID)

		templateData = map[string]interface{}{
			"User1Post1ConversationID":  user1Post1ConversationID,
			"User1Post2ConversationID":  user1Post2ConversationID,
			"User2Post1ConversationID":  user2Post1ConversationID,
			"User2Post1Conversation2ID": user2Post1Conversation2ID,
			"User3Post1ConversationID":  user3Post1ConversationID,

			"User1Post1ID": user1Post1ID,
			"User1Post2ID": user1Post2ID,
			"User2Post1ID": user2Post1ID,
			"User1Post3ID": user1Post3ID,
			"User3Post1ID": user3Post1ID,

			"User1ID": user1ID,
			"User2ID": user2ID,
			"User3ID": user3ID,
		}
	}

	for _, args := range []string{"", "input: {}"} {
		listConversationsQuery := fmt.Sprintf(`{
			conversations(%s) {
				conversations {
					id
					post {
						id
					}
					startedBy {
						id
					}
					participants {
						id
					}
				}
			}
		}`, args)

		t.Run("conversations "+args, func(t *testing.T) {
			t.Run(args, func(t *testing.T) {
				t.Run("user1 sees conversations with user2", func(t *testing.T) {
					harness := NewTestHarness(t)

					harness.GQLAssert("", GQLAssertInput{
						ExecInput: ExecInput{
							UserID: user1IDi64,
							Query:  listConversationsQuery,
						},
						ExpectedResult: harness.TemplateString(`
							{
								"conversations": {
									"conversations": [
										{
											"id": "[[ .User1Post1ConversationID ]]",
											"participants": [
												{
													"id": "[[ .User1ID ]]"
												},
												{
													"id": "[[ .User2ID ]]"
												}
											],
											"post": {
												"id": "[[ .User1Post1ID ]]"
											},
											"startedBy": {
												"id": "[[ .User2ID ]]"
											}
										},
										{
											"id": "[[ .User1Post2ConversationID ]]",
											"participants": [
												{
													"id": "[[ .User1ID ]]"
												},
												{
													"id": "[[ .User2ID ]]"
												}
											],
											"post": {
												"id": "[[ .User1Post2ID ]]"
											},
											"startedBy": {
												"id": "[[ .User2ID ]]"
											}
										},
										{
											"id": "[[ .User2Post1ConversationID ]]",
											"participants": [
												{
													"id": "[[ .User1ID ]]"
												},
												{
													"id": "[[ .User2ID ]]"
												}
											],
											"post": {
												"id": "[[ .User2Post1ID ]]"
											},
											"startedBy": {
												"id": "[[ .User1ID ]]"
											}
										}
									]
								}
							}`, templateData),
					})
				})

				// proves that user1 isnt just seeing all conversations
				t.Run("user3 has conversations with user2", func(t *testing.T) {
					harness := NewTestHarness(t)

					harness.GQLAssert("", GQLAssertInput{
						ExecInput: ExecInput{
							UserID: user3IDi64,
							Query:  listConversationsQuery,
						},
						ExpectedResult: harness.TemplateString(`
							{
								"conversations": {
									"conversations": [
										{
											"id": "[[ .User2Post1Conversation2ID ]]",
											"participants": [
												{
													"id": "[[ .User2ID ]]"
												},
												{
													"id": "[[ .User3ID ]]"
												}
											],
											"post": {
												"id": "[[ .User2Post1ID ]]"
											},
											"startedBy": {
												"id": "[[ .User3ID ]]"
											}
										},
										{
											"id": "[[ .User3Post1ConversationID ]]",
											"participants": [
												{
													"id": "[[ .User2ID ]]"
												},
												{
													"id": "[[ .User3ID ]]"
												}
											],
											"post": {
												"id": "[[ .User3Post1ID ]]"
											},
											"startedBy": {
												"id": "[[ .User2ID ]]"
											}
										}
									]
								}
							}`, templateData),
					})
				})
			})
		})
	}

	t.Run("list conversations by post id", func(t *testing.T) {
		postID := user2Post1ID
		query := fmt.Sprintf(`{
			conversations(input: {postID: "%s"}) {
				conversations {
					id
				}
			}
		}`, postID)

		t.Run("user1 sees conversation with user2 about user2post1", func(t *testing.T) {
			harness.GQLAssert("", GQLAssertInput{
				ExecInput: ExecInput{
					UserID: user1IDi64,
					Variables: map[string]interface{}{
						"input": map[string]interface{}{
							"postID": postID,
						},
					},
					Query: query,
				},
				ExpectedResult: harness.TemplateString(`
					{
						"conversations": {
							"conversations": [
								{
									"id": "[[ .User2Post1ConversationID ]]"
								}
							]
						}
					}`, templateData),
			})
		})

		t.Run("user2 sees conversation with user1 and user3 and user2post1", func(t *testing.T) {
			harness.GQLAssert("", GQLAssertInput{
				ExecInput: ExecInput{
					UserID: user2IDi64,
					Query:  query,
				},
				ExpectedResult: harness.TemplateString(`
					{
						"conversations": {
							"conversations": [
								{
									"id": "[[ .User2Post1ConversationID ]]"
								},
								{
									"id": "[[ .User2Post1Conversation2ID ]]"
								}
							]
						}
					}`, templateData),
			})
		})
	})

	t.Run("get my conversation by id", func(t *testing.T) {
		harness.GQLAssert("", GQLAssertInput{
			ExecInput: ExecInput{
				UserID: user1IDi64,
				Query: fmt.Sprintf(`{
					conversationByID(id: "%s") {
						id
					}
				}`, user1Post1ConversationID),
			},
			ExpectedResult: harness.TemplateString(`
				{
					"conversationByID": {
						"id": "[[ .User1Post1ConversationID ]]"
					}
				}`, templateData),
		})
	})

	t.Run("get another user's conversation by id", func(t *testing.T) {
		harness.GQLAssert("", GQLAssertInput{
			ExecInput: ExecInput{
				UserID: user3IDi64,
				Query: fmt.Sprintf(`{
					conversationByID(id: "%s") {
						id
					}
				}`, user1Post1ConversationID),
			},
			ExpectedErrors: []*errors.QueryError{
				{
					Message: "unauthorized",
				},
			},
		})
	})

	t.Run("send and get messages", func(t *testing.T) {
		// send messages
		{
			harness := NewTestHarness(t)
			sendMessage := func(userID int64, conversationID string, body string) {
				harness.GQLAssert("", GQLAssertInput{
					ExecInput: ExecInput{
						UserID: userID,
						Variables: map[string]interface{}{
							"input": map[string]interface{}{
								"conversationID": conversationID,
								"body":           body,
							},
						},
						Query: `mutation SendMessage($input: SendMessageInput!) {
							sendMessage(input: $input) {
								message {
									body
								}
							}
						}`,
					},
					ExpectedResult: fmt.Sprintf(`
					{
						"sendMessage": {
							"message": {
								"body": "%s"
							}
						}
					}
					`, body),
				})
			}

			sendMessage(user1IDi64, user2Post1ConversationID, "body 1")
			sendMessage(user2IDi64, user2Post1ConversationID, "body 2")
			sendMessage(user1IDi64, user2Post1ConversationID, "body 3")

			// This had a bug (NPE) before so testing that we can
			// get messages for a created conversation
			rawRes := map[string]interface{}{}
			harness.MustExec(ExecInput{
				UserID: user1IDi64,
				Query: fmt.Sprintf(`
				mutation {
					getOrCreateConversation(input: {
						postID: "%s"
					}) {
						conversation {
							id
							messages() {
								messages {
									from {
										id
									}
								}
							}
						}
					}
				 }`, user2Post1ID),
			}, &rawRes)
		}

		// both users should see the same thing since it's one conversation
		for _, userID := range []int64{user1IDi64, user2IDi64} {
			t.Run(fmt.Sprintf("get messages for user %d", userID), func(t *testing.T) {
				harness := NewTestHarness(t)
				harness.GQLAssert("", GQLAssertInput{
					ExecInput: ExecInput{
						UserID: userID,
						Query: harness.TemplateString(`{
							conversationByID(id: "[[ .User2Post1ConversationID ]]") {
								messages() {
									messages {
										body
									}
								}
							}
						}`, templateData),
					},
					ExpectedResult: harness.TemplateString(`
					{
						"conversationByID": {
							"messages": {
								"messages": [
									{
										"body": "body 3"
									},
									{
										"body": "body 2"
									},
									{
										"body": "body 1"
									}
								]
							}
						}
					}`, templateData),
				})
			})

			t.Run(fmt.Sprintf("paginate messages for user %d", userID), func(t *testing.T) {
				getBody := func(res map[string]interface{}, index int) string {
					message := res["messages"].([]interface{})[index].(map[string]interface{})
					body := message["body"].(string)
					return body
				}

				// page 1
				var rawRes map[string]interface{}
				harness.MustExec(ExecInput{
					UserID: userID,
					Query: fmt.Sprintf(`{
						conversationByID(id: "%s") {
							messages(input: {limit: 2}) {
								messages {
									body
								}

								nextPageToken
							}
						}
					}`, user2Post1ConversationID),
				}, &rawRes)

				res := rawRes["conversationByID"].(map[string]interface{})["messages"].(map[string]interface{})
				require.NotNil(t, res["nextPageToken"])
				nextPageToken := res["nextPageToken"].(string)
				require.NotEmpty(t, nextPageToken)
				assert.Equal(t, "body 3", getBody(res, 0))
				assert.Equal(t, "body 2", getBody(res, 1))

				rawRes = map[string]interface{}{}
				harness.MustExec(ExecInput{
					UserID: userID,
					Query: fmt.Sprintf(`{
						conversationByID(id: "%s") {
							messages(input: {limit: 2, pageToken: "%s"}) {
								messages {
									body
								}

								nextPageToken
							}
						}
					}`, user2Post1ConversationID, nextPageToken),
				}, &rawRes)

				res = rawRes["conversationByID"].(map[string]interface{})["messages"].(map[string]interface{})
				assert.Nil(t, res["nextPageToken"])
				assert.Equal(t, "body 1", getBody(res, 0))
			})
		}
	})

	t.Run("send messages to another user's conversation", func(t *testing.T) {
		harness.GQLAssert("", GQLAssertInput{
			ExecInput: ExecInput{
				UserID: user1IDi64,
				Variables: map[string]interface{}{
					"input": map[string]interface{}{
						"conversationID": user3Post1ConversationID,
						"body":           "doesnt matter",
					},
				},
				Query: `
				mutation SendMessage($input: SendMessageInput!) {
					sendMessage(input: $input) {
						message {
							body
						}
					}
				}
				`,
			},
			ExpectedErrors: []*errors.QueryError{
				{
					Message: "unauthorized",
				},
			},
		})
	})

	// used to return unauthorized
	t.Run("get messages for conversation", func(t *testing.T) {
		rawRes := map[string]interface{}{}
		harness.MustExec(ExecInput{
			UserID: user1IDi64,
			Query: `{
				conversations() {
					conversations {
						messages() {
							messages {
								body
							}
						}
					}
				}
			}`,
		}, &rawRes)
	})
}
