package gqlschema

const typeDefs = `
	scalar Timestamp

	schema {
		query: Query
		mutation: Mutation
	}

	type Query {
		hello(): String!
		currentUser(): User
		otherUser(id: Int!): User
		
		// Who are my followers
		getFollowersByUserID(id: Int!): FollowersResult!

		// Who am I following
		getFollowingByUserID(id: Int!): FollowersResult!

		getCommentsByPostID(id: Int!): PostCommentsResult!

		feed(input: FeedInput): FeedResult
		conversationByID(id: String!): Conversation!
		conversations(input: ConversationsInput): ConversationsResult!
		notifications(input: NotificationsInput): NotificationsResult!

		hasCurrentUserLikedPost(id: Int!): Boolean!
		isLiked(id: Int!): Boolean!
		postComments(input: UserPostCommentsInput!): PostCommentsResult!
	}

	type Mutation {
		createPost(input: CreatePostInput!): Post
		updatePost(input: UpdatePostInput!): Post
		removePost(input: RemovePostInput!): Post

		// post view times
		postViewTimes(postID: String!): String

		// Post Comments
		createPostComment(input: CreatePostCommentInput!): Boolean!
		removePostComment(input: RemovePostCommentInput!): Boolean!

		createReportedPost(input: ReportedPostInput!): ReportedPost
		createFollower(id: Int!): Follower
		unfollow(id: Int!): Boolean!

		updateUser(input: UpdateUserInput!): User
		updateFCMToken(input: UpdateTokenInput!): User

		getOrCreateConversation(input: GetOrCreateConversationInput!): GetOrCreateConversationResult!
		sendMessage(input: SendMessageInput!): SendMessageResult

		userAssignDeviceToken(input: UserAssignDeviceTokenInput!): Boolean

		requestLoginCode(input: RequestLoginCodeInput!): Boolean
		loginUser(input: LoginUserInput!): LoginUserResult

		markNotificationRead(input: MarkNotificationReadInput!): Boolean

		requestMediaUpload(input: RequestMediaUploadInput!): RequestMediaUploadResult

		likePost(id: Int!): Boolean!
		unlikePost(id: Int!): Boolean!
	}

	input RemovePostInput {
		id: ID!
	}

	input ReportedPostInput {
		postID: ID!
		personReportingID: ID!
		ReportingReasonText: String!
	}

	input NotificationsInput {
		keepSchemaHappy: Boolean
	}

	type Follower {
		id: ID!
		userID: Int!
		followerUserID: Int!
	}

	type NotificationsResult {
		notifications: [Notification]!
	}

	type Notification {
		id: ID!
		unread: Boolean!
		content: String!
		timestamp: Timestamp!
	}

	input MarkNotificationReadInput {
		notificationID: ID!
	}

	input MessagesInput {
		after: ID!
		before: ID!
	}

	enum PostKind {
		TEXT
		IMAGE
		VIDEO
	}

	input CreatePostInput {
		title: String!
		description: String
		
		// kind specifies text/image/video 
		kind: PostKind!

		poster: String

		mediaKind: String
		mediaURL: String
		mediaMetadata: MediaMetadataInput

		tags: [String!]
	}

	input CreatePostCommentInput {
		parentCommentID: Int = 0
		comment: String!
		postID: Int!
	}

	input RemovePostCommentInput {
		commentID: Int!
		postID: Int!
	}

	input MediaMetadataInput {
		width: Int
		height: Int
	}

	input UserAssignDeviceTokenInput {
		deviceToken: String!
	}

	input UpdatePostInput {
		id: ID!
		title: String
		description: String
		poster: String
		tags: [String!]
	}

	input LoginUserInput {
		phoneNumber: String!
		loginCode: String!
		fcmToken: String
	}

	type LoginUserResult {
		token: String!
		expiresAt: Timestamp!
	}

	
	input RequestLoginCodeInput {
		phoneNumber: String!
	}

	// type RequestUserLoginResult {
	// 	phoneNumber: String
	// 	expiresAt: String
	// }

	input UpdateUserInput {
		name: String
		photoURL: String
		zipCode: String
		bio: String
	}

	input UpdateTokenInput {
		userID: ID!
		fcmToken: String!
	}

	type Post {
		id: ID!

		author: User

		neighborhood: Neighborhood

		title: String!
		description: String
		kind: PostKind

		processing: Boolean!

		tags: [String!]

		poster: String

		shareLinkURL: String

		preview: PostMedia
		media: PostMedia
		
		// relatedPosts: [Post]!

		createdAt: Timestamp!
		updatedAt: Timestamp!

		Likes: Int
		commentCount: Int
		viewTimes: String
	}

	type PostMedia {
		url: String
		width: Int
		height: Int
	}

	type ReportedPost {
		id: ID!
		postID: Int!
		personReportingID: Int!
		reportingReasonText: String!
		ActionTaken: Int!

		createdAt: Timestamp!
		updatedAt: Timestamp!
	}

	type Neighborhood {
		id: ID!
		slug: String!
		name: String!
	}

	type User {
		id: ID!

		phoneNumber: String
		name: String
		photoURL: String
		zipCode: String
		bio: String
		followers: Int
		following: Int

		postCount: Int!
		
		createdAt: Timestamp!
		updatedAt: Timestamp!
		
		posts(input: UserPostsInput): UserPostsResult!
	}

	input FeedInput {
		tags: [String!]

		pageToken: String
		limit: Int
	}

	type FeedResult {
		posts: [Post!]!
		nextPageToken: String
	}

	enum MediaType {
		IMAGE
		VIDEO
		AVATAR
	}

	input RequestMediaUploadInput {
		mediaType: MediaType!
	}

	type RequestMediaUploadResult {
		// putURL is the presigned S3 URL you upload file to using HTTP PUT
		putURL: String!
		// getURL is the (public) URL you can send an HTTP GET to for
		// the image. you can set this as mediaURL in post
		getURL: String!
	}

	input UserPostsInput {
		// default always backwards (reverse chronological), it's a social network
		// orderBy: OrderPostsBy
		pageToken: String
		limit: Int
		otherUserID: Int
	}

	type UserPostsResult {
		posts: [Post!]!
		nextPageToken: String
	}

	/*
		Post Comments
	*/
	input UserPostCommentsInput {
		pageToken: String
		limit: Int
		// otherUserID: Int
		postID: Int!
	}

	/*
		Chat
	*/

	type Conversation {
		id: ID!

		// Keep post nullable since we may not tie conversations to posts
		// in the future. Client should filter non null for now.
		post: Post

		startedBy: User!
		createdAt: Timestamp!
		participants: [User!]!

		messages(input: ConversationMessagesInput): ConversationMessagesResult!
	}

	type Message {
		id: ID!

		from: User!
		body: String!

		timestamp: Timestamp!
	}

	input GetOrCreateConversationInput {
		postID: ID!
	}

	type GetOrCreateConversationResult {
		conversation: Conversation!
		created: Boolean!
	}

	input SendMessageInput {
		conversationID: ID!
		destinationUserID: ID!
		body: String!
	}

	type SendMessageResult {
		message: Message!
	}

	input ConversationsInput {
		postID: ID

		// TODO on server, but client should implement
		pageToken: String
		limit: Int
	}

	type ConversationsResult {
		conversations: [Conversation!]!
		nextPageToken: String
	}

	type FollowersResult {
		followers: [User!]!
	}

	input ConversationMessagesInput {
		pageToken: String
		limit: Int
	}

	type ConversationMessagesResult {
		messages: [Message!]!
		nextPageToken: String
	}

	type PostCommentsResult {
		comments: [PostComment!]!
		nextPageToken: String
	}

	type PostComment {
		id: ID!
		// parentCommentID: ID!
		comment: String!
		author: User!

		createdAt: Timestamp!
		updatedAt: Timestamp!
	}
`
