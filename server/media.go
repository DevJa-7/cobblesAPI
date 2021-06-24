package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediaconvert"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// var s3VideoURLRegex = regexp.MustCompile(`videos\/([\w-]+)`)

type JobDetails struct {
	Timestamp    int64  `json:"timestamp"`
	AccountID    string `json:"accountId"`
	Queue        string `json:"queue"`
	JobID        string `json:"jobId"`
	Status       string `json:"status"`
	UserMetadata struct {
		PostIDStr string `json:"post_id_str" mapstructure:"post_id_str"`
	} `json:"userMetadata"`
	OutputGroupDetails []struct {
		OutputDetails []struct {
			OutputFilePaths []string `json:"outputFilePaths"`
			DurationInMs    int      `json:"durationInMs"`
			VideoDetails    struct {
				WidthInPx  int `json:"widthInPx"`
				HeightInPx int `json:"heightInPx"`
			} `json:"videoDetails"`
		} `json:"outputDetails"`
		Type              string   `json:"type"`
		PlaylistFilePaths []string `json:"playlistFilePaths,omitempty"`
	} `json:"outputGroupDetails"`
}

type PostMedia struct {
	URL    string `json:"url,omitempty"`
	Width  int32  `json:"width,omitempty"`
	Height int32  `json:"height,omitempty"`
}

func (s *Server) ProcessPostMediaUpload(postID int64, uploadedMediaURL string) error {
	uu, err := url.Parse(uploadedMediaURL)
	if err != nil {
		return err
	}

	var bucketName string
	parts := strings.Split(uu.Host, ".")
	if len(parts) > 0 {
		bucketName = parts[0]
	}

	s3Path := fmt.Sprintf("s3://%s%s", bucketName, uu.Path)
	log.Println(s3Path)

	_, err = s.MediaConvert.CreateJob(&mediaconvert.CreateJobInput{
		JobTemplate: aws.String("process-video"),

		Settings: &mediaconvert.JobSettings{
			Inputs: []*mediaconvert.Input{
				&mediaconvert.Input{
					FileInput: aws.String(s3Path),
				},
			},
		},

		UserMetadata: map[string]*string{
			"post_id_str": aws.String(strconv.FormatInt(postID, 10)),
		},

		Role: aws.String("arn:aws:iam::927717636424:role/cobbles-mediaconvert-s3"),
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) ProcessMediaQueue() error {
	for {
		output, err := s.SQS.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl: aws.String(s.SQSMediaProcessingQueueURL),
		})
		if err != nil {
			return err
		}

		for _, message := range output.Messages {
			if message.Body != nil {
				var m map[string]interface{}
				if err := json.Unmarshal([]byte(*message.Body), &m); err != nil {
					return err
				}

				var jobDetails JobDetails
				if err := mapstructure.Decode(m["data"], &jobDetails); err != nil {
					return err
				}

				postID, err := strconv.ParseInt(jobDetails.UserMetadata.PostIDStr, 10, 64)
				if err != nil {
					log.Println(err)
					continue
				}

				var previewMedia PostMedia
				var media PostMedia
				switch jobDetails.Status {
				case "COMPLETE":
					for _, outputGroupDetails := range jobDetails.OutputGroupDetails {
						for _, outputDetail := range outputGroupDetails.OutputDetails {
							width := int32(outputDetail.VideoDetails.WidthInPx)
							height := int32(outputDetail.VideoDetails.HeightInPx)

							outputURL, _ := url.Parse(outputDetail.OutputFilePaths[0])
							mediaURL := fmt.Sprintf("https://%s.s3-external-1.amazonaws.com%s", outputURL.Host, outputURL.Path)

							switch outputGroupDetails.Type {
							case "FILE_GROUP":
								previewMedia.Width = width
								previewMedia.Height = height
								previewMedia.URL = mediaURL
							case "HLS_GROUP":
								media.Width = width
								media.Height = height
								media.URL = mediaURL
							}
						}
					}

					_, err = s.ConnPool.Exec(`
						update posts
						set processing = false, media = $2, preview = $3
						where id = $1
					`, postID, media, previewMedia)
					if err != nil {
						return err
					}

					s.SQS.DeleteMessage(&sqs.DeleteMessageInput{
						QueueUrl:      aws.String(s.SQSMediaProcessingQueueURL),
						ReceiptHandle: message.ReceiptHandle,
					})
				default:
					s.SQS.DeleteMessage(&sqs.DeleteMessageInput{
						QueueUrl:      aws.String(s.SQSMediaProcessingQueueURL),
						ReceiptHandle: message.ReceiptHandle,
					})
				}
			}

		}
	}
}
