package server

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/mediaconvert"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/log/zerologadapter"
	"github.com/jinzhu/gorm"

	// Postgres
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/rs/zerolog"
)

// Server ...
type Server struct {
	S3ProcessedMediaBucket     string
	S3UserMediaBucket          string
	S3ImageProxyBaseURL        string
	S3ProcessedUserMediaBucket string
	S3                         *s3.S3

	SQSMediaProcessingQueueURL string
	SNSPlatformApplicationArn  string

	ImgixProcessedMediaEndpoint string
	ImgixUserMediaMediaEndpoint string

	SQS          *sqs.SQS
	SNS          *sns.SNS
	ConnPool     *pgx.ConnPool
	DB           *gorm.DB
	MediaConvert *mediaconvert.MediaConvert

	ServerSecret string
}

// NewServer ...
func NewServer() *Server {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("specify DATABASE_URL")
	}

	connConfig, err := pgx.ParseConnectionString(databaseURL)
	if err != nil {
		log.Fatal(err)
	}

	logger := zerolog.New(os.Stdout)
	connConfig.LogLevel = pgx.LogLevelDebug
	connConfig.Logger = zerologadapter.NewLogger(logger)

	// Legacy method of connecting db.
	// We'll use GORM for new codes
	connPool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     connConfig,
		MaxConnections: 10000,
	})
	if err != nil {
		log.Fatal(err)
	}

	// GORM config
	db, err := gorm.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	if err := db.DB().Ping(); err != nil {
		log.Fatal(err)
		return nil
	}

	// defer db.Close()
	db.DB().SetMaxOpenConns(10000)
	db.DB().SetMaxIdleConns(5000)

	// Auto migration
	db.AutoMigrate(
		&User{},
		&Post{},
		&ReportedPost{},
		&Follower{},
		&Like{},
		&PostComment{},
	)

	// many many AWS services
	sess := session.New(&aws.Config{
		Region: aws.String(endpoints.UsEast1RegionID),
	})

	sns := sns.New(sess)
	s3 := s3.New(sess)
	sqs := sqs.New(sess)
	mediaConvert := mediaconvert.New(sess, &aws.Config{
		Endpoint: aws.String("https://vasjpylpa.mediaconvert.us-east-1.amazonaws.com"),
		// Endpoint: aws.String("https://8lrsp3p0.mediaconvert.us-east-1.amazonaws.com"),
	})

	serverSecret := os.Getenv("SERVER_SECRET")
	if serverSecret == "" {
		log.Fatal("set SERVER_SECRET")
	}

	s3MediaBucket := os.Getenv("S3_MEDIA_BUCKET")
	if s3MediaBucket == "" {
		s3MediaBucket = "llc-cobbles-dev-user-media"
	}

	s3ProcessedMediaBucket := os.Getenv("S3_PROCESSED_MEDIA_BUCKET")
	if s3ProcessedMediaBucket == "" {
		s3ProcessedMediaBucket = "llc-cobbles-dev-processed-user-media"
	}

	s3ImageProxyBaseURL := os.Getenv("S3_IMAGE_PROXY_BASE_URL")
	if s3ImageProxyBaseURL == "" {
		s3ImageProxyBaseURL = "https://llc-cobbles-dev-user-images.imgix.net"
	}

	snsApplicationArn := os.Getenv("SNS_APP_ARN")
	if snsApplicationArn == "" {
		// snsApplicationArn = "arn:aws:sns:us-east-1:236073164598:app/APNS_SANDBOX/cobbles-dev"
		snsApplicationArn = "arn:aws:sns:us-east-1:927717636424:app/GCM/cobbles-dev"
	}

	sqsMediaProcessingQueueURL := os.Getenv("SQS_MEDIA_PROCESSING_QUEUE_URL")
	if sqsMediaProcessingQueueURL == "" {
		sqsMediaProcessingQueueURL = "https://sqs.us-east-1.amazonaws.com/927717636424/cobbles_media_events.fifo"
		//sqsMediaProcessingQueueURL = "https://sqs.us-east-1.amazonaws.com/236073164598/cobbles_media_events.fifo"
	}

	return &Server{
		S3ProcessedMediaBucket:     s3ProcessedMediaBucket,
		S3UserMediaBucket:          s3MediaBucket,
		S3ImageProxyBaseURL:        s3ImageProxyBaseURL,
		S3:                         s3,
		SQSMediaProcessingQueueURL: sqsMediaProcessingQueueURL,
		SQS:                        sqs,
		SNSPlatformApplicationArn:  snsApplicationArn,

		// ImgixProcessedMediaEndpoint: "https://processed-user-media.imgix.net",
		// ImgixUserMediaMediaEndpoint: "https://lc-cobbles-dev-user-images.imgix.net",
		ImgixProcessedMediaEndpoint: "https://llc-cobbles-dev-processed-user-images.imgix.net",
		ImgixUserMediaMediaEndpoint: "https://llc-cobbles-dev-user-images.imgix.net",

		ConnPool:     connPool, // Legacy
		DB:           db,
		SNS:          sns,
		MediaConvert: mediaConvert,
		ServerSecret: serverSecret,
	}
}
